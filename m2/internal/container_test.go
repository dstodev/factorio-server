package internal_test

import (
	"errors"
	"testing"

	"manage2/internal"
	"manage2/internal/program"
	"manage2/internal/stream"
)

const commonImage = "alpine:latest"

func TestContainer(t *testing.T) {
	image := "some-image:latest"
	c := internal.NewContainer(image)

	if c.Image != image {
		t.Fatalf("expected image '%s', got: %s", image, c.Image)
	}

	if err := c.Run(); !errors.Is(err, internal.ErrNoProgram) {
		t.Fatalf("expected error '%v', got: %v", internal.ErrNoProgram, err)
	}
	if c.ID != "" {
		t.Fatalf("expected empty ID, got: %s", c.ID)
	}
}

func TestContainerWithProgram(t *testing.T) {
	c := internal.NewContainer(commonImage)
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "exit 5"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	AssertResult(t, result, c.ID, 5, nil)
}

func AssertResult(t *testing.T, result internal.ContainerResult, expectedID string, expectedStatus int, expectedErr error) {
	t.Helper()
	if expectedID == "!" {
		if result.ID == "" {
			t.Errorf("expected non-empty ID, got: %s", result.ID)
		}
	} else {
		if result.ID != expectedID {
			t.Errorf("expected ID '%s', got: %s", expectedID, result.ID)
		}
	}
	if result.Status != expectedStatus {
		t.Errorf("expected exit code %d, got: %d", expectedStatus, result.Status)
	}
	if !errors.Is(result.Err, expectedErr) {
		t.Errorf("expected error '%v', got: %v", expectedErr, result.Err)
	}
}

func TestContainerWithStdout(t *testing.T) {
	stdoutChan := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(stdoutChan))
	p, err := program.FromSystem("printf", program.WithStringArgs("|%s|\n", "Hello,", "world!"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	AssertResult(t, result, c.ID, 0, nil)

	expectedOutput := "|Hello,|\n|world!|\n|--|\n"
	select {
	case msg := <-stdoutChan:
		if msg != expectedOutput {
			t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}
}

func TestContainerWithStdoutAndStderr(t *testing.T) {
	stdoutChan := make(chan string, 1)
	stderrChan := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(stdoutChan),
		internal.WithStderrChannel(stderrChan))
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "printf '|%s|\n' Hello, world! | tee /dev/stderr"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	expectedOutput := "|Hello,|\n|world!|\n"

	result := <-c.Done
	AssertResult(t, result, c.ID, 0, nil)

	select {
	case msg := <-stdoutChan:
		if msg != expectedOutput {
			t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}

	select {
	case msg := <-stderrChan:
		if msg != expectedOutput {
			t.Fatalf("expected stderr message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Fatal("expected a message on stderr")
	}
}

func TestContainerWithStdin(t *testing.T) {
	stdinChan := make(chan string, 1)
	stdoutChan := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stdinChan),
		internal.WithStdoutChannel(stdoutChan))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && echo \"|$input|\""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msg := "Hello, world!\n"
	stdinChan <- msg
	close(stdinChan)

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	AssertResult(t, result, c.ID, 0, nil)

	expectedMsg := "|Hello, world!|\n"
	select {
	case msg := <-stdoutChan:
		if msg != expectedMsg {
			t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}
}

func TestContainerStdinOnly(t *testing.T) {
	stdinChan := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stdinChan))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msg := "5\n"
	stdinChan <- msg
	close(stdinChan)

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	AssertResult(t, result, c.ID, 5, nil)
}

func TestBuildProgramCommandStrings(t *testing.T) {
	c := internal.NewContainer(commonImage)
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "echo Hello"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cmdTokens, mounts := c.BuildProgramCmd(p)

	expectedCmd := []string{"/bin/sh", "-c", "echo Hello", "--"}
	if len(cmdTokens) != len(expectedCmd) {
		t.Fatalf("expected command tokens %v, got: %v", expectedCmd, cmdTokens)
	}
	for i, token := range cmdTokens {
		if token != expectedCmd[i] {
			t.Fatalf("expected command token '%s', got: '%s'", expectedCmd[i],
				token)
		}
	}
	if len(mounts) != 0 {
		t.Fatalf("expected no mounts, got: %v", mounts)
	}
}

func TestBuildProgramCommandWithFileArgs(t *testing.T) {
	dir := internal.NewTestDir(t)
	script := dir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"\n", 0755)
	expectedMsg := "Hello, world!\n"
	file := dir.WriteFile("file.txt", expectedMsg, 0644)

	stdoutChan := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(stdoutChan))
	p, err := program.FromFile(script,
		program.WithFileArgs(file))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cmdTokens, fileMap := c.BuildProgramCmd(p)

	if len(fileMap) != 2 { // script & file
		t.Fatalf("expected 2 mounts, got: %d", len(fileMap))
	}

	expectedCmd := []string{fileMap[script], fileMap[file], "--"}
	if len(cmdTokens) != len(expectedCmd) {
		t.Fatalf("expected command tokens %v, got: %v", expectedCmd, cmdTokens)
	}

	for i, token := range cmdTokens {
		if token != expectedCmd[i] {
			t.Fatalf("expected command token '%s', got: '%s'", expectedCmd[i], token)
		}
	}

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	AssertResult(t, result, c.ID, 0, nil)

	select {
	case msg := <-stdoutChan:
		if msg != expectedMsg {
			t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}
}

// Assert we can control when a container exits, by e.g. running a command that
// waits for input on stdin like `cat`, then explicitly allowing the container
// to close later by closing the stdin channel. This is useful to run multiple
// arbitrary commands with Exec() before closing the container.
func TestIdleContainer(t *testing.T) {
	stopWait := make(chan string, 1)
	c := internal.NewContainer(commonImage, internal.WithStdinChannel(stopWait))
	p, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	close(stopWait) // Test times out without closing stdin; container is idle

	result := <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}

func TestExec(t *testing.T) {
	stdinChan := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stdinChan))
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	execChan := make(chan string, 1)

	resultChan := c.Exec(p, stream.WithStdinChannel(execChan))

	execChan <- "4\n"
	close(execChan)

	result := <-resultChan
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 4 {
		t.Fatalf("expected exit code 4, got: %d", result.Status)
	}

	stdinChan <- "5\n"
	close(stdinChan)

	result = <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 5 {
		t.Fatalf("expected exit code 5, got: %d", result.Status)
	}
}

func TestExecNotRunning(t *testing.T) {
	c := internal.NewContainer(commonImage)
	p, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	resultChan := c.Exec(p)

	result := <-resultChan
	AssertResult(t, result, "", -1, internal.ErrNotRunning)
}

func TestExecNoPrograms(t *testing.T) {
	c := internal.NewContainer(commonImage)

	resultChan := c.Exec(nil)

	result := <-resultChan
	AssertResult(t, result, "", -1, internal.ErrNoProgram)
}

func TestExecMissingMounts(t *testing.T) {
	stopWait := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stopWait))
	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(wait); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	path := internal.NewTestDir(t).TouchFile("file.txt")
	p, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	resultChan := c.Exec(p)

	// Must wait for exec program to finish before closing the container!
	result := <-resultChan
	AssertResult(t, result, "", -1, internal.ErrNotMounted)

	close(stopWait) // Allow container to close

	result = <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}

func TestExecMounts(t *testing.T) {
	stopWait := make(chan string, 1)
	path := internal.NewTestDir(t).TouchFile("file.txt")
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stopWait),
		internal.WithMounts(path))

	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := c.Run(wait); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	p, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	resultChan := c.Exec(p)

	// Must wait for exec program to finish before closing the container!
	result := <-resultChan
	AssertResult(t, result, "!", 0, nil)

	close(stopWait) // Allow container to close

	result = <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}

func TestExecMultiWrite(t *testing.T) {
	path := internal.NewTestDir(t).WriteFile(
		"file.txt",
		`#!/bin/sh
while :; do
	read line
	if [ -z "$line" ]; then
		break
	fi
	echo "$line"
done
`, 0744)
	stopWait := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stopWait),
		internal.WithMounts(path))

	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := c.Run(wait); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, err := program.FromFile(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	stdinChan := make(chan string, 1)
	stdoutChan := make(chan string, 1)

	resultChan := c.Exec(p,
		stream.WithStdinChannel(stdinChan),
		stream.WithStdoutChannel(stdoutChan))

	stdinChan <- "Hello,\n"
	msg := <-stdoutChan
	if msg != "Hello,\n" {
		t.Fatalf("expected stdout message 'Hello,', got: %s", msg)
	}

	stdinChan <- "world!\n"
	msg = <-stdoutChan
	if msg != "world!\n" {
		t.Fatalf("expected stdout message 'world!', got: %s", msg)
	}

	close(stdinChan) // Signal end of input

	result := <-resultChan
	AssertResult(t, result, "!", 0, nil)

	close(stopWait) // Allow container to close

	msg, ok := <-stdoutChan
	if ok {
		t.Fatalf("expected no message on stdout after closing stdin, got: %s", msg)
	}
	if msg != "" {
		t.Fatalf("expected empty message on stdout, got: %s", msg)
	}

	result = <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}
