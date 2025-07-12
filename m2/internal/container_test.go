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
	ctrStdout := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(ctrStdout))
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
	case msg := <-ctrStdout:
		if msg != expectedOutput {
			t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}
}

func TestContainerWithStdoutAndStderr(t *testing.T) {
	ctrStdout := make(chan string, 1)
	ctrStderr := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(ctrStdout),
		internal.WithStderrChannel(ctrStderr))
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
	case msg := <-ctrStdout:
		if msg != expectedOutput {
			t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}

	select {
	case msg := <-ctrStderr:
		if msg != expectedOutput {
			t.Fatalf("expected stderr message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Fatal("expected a message on stderr")
	}
}

func TestContainerWithStdin(t *testing.T) {
	ctrStdin := make(chan string, 1)
	ctrStdout := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin),
		internal.WithStdoutChannel(ctrStdout))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && echo \"|$input|\""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msg := "Hello, world!\n"
	ctrStdin <- msg
	close(ctrStdin)

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
	case msg := <-ctrStdout:
		if msg != expectedMsg {
			t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}
}

func TestContainerStdinOnly(t *testing.T) {
	ctrStdin := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msg := "5\n"
	ctrStdin <- msg
	close(ctrStdin)

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
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "echo Hello"))
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

	ctrStdout := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(ctrStdout))
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
	case msg := <-ctrStdout:
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
	ctrStdin := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin))
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

	close(ctrStdin) // Test times out without closing stdin; container is idle

	result := <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}

func TestExec(t *testing.T) {
	ctrStdin := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin))
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

	execStdin := make(chan string, 1)

	execResult := c.Exec(p, stream.WithStdinChannel(execStdin))

	select {
	case result := <-execResult:
		t.Fatalf("expected no result, got: %v", result.Err)
	default:
	}

	execStdin <- "4\n"
	close(execStdin)

	result := <-execResult
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 4 {
		t.Fatalf("expected exit code 4, got: %d", result.Status)
	}

	ctrStdin <- "5\n"
	close(ctrStdin)

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
	execResult := c.Exec(p)

	result := <-execResult
	AssertResult(t, result, "", -1, internal.ErrNotRunning)
}

func TestExecNoPrograms(t *testing.T) {
	c := internal.NewContainer(commonImage)

	execResult := c.Exec(nil)

	result := <-execResult
	AssertResult(t, result, "", -1, internal.ErrNoProgram)
}

func TestExecMissingMounts(t *testing.T) {
	ctrStdin := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin))
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
	execResult := c.Exec(p)

	// Must wait for exec program to finish before closing the container!
	result := <-execResult
	AssertResult(t, result, "", -1, internal.ErrNotMounted)

	close(ctrStdin) // Allow container to close

	result = <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}

func TestExecMounts(t *testing.T) {
	ctrStdin := make(chan string, 1)
	path := internal.NewTestDir(t).TouchFile("file.txt")
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin),
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
	execResult := c.Exec(p)

	// Must wait for exec program to finish before closing the container!
	result := <-execResult
	AssertResult(t, result, "!", 0, nil)

	close(ctrStdin) // Allow container to close

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
	ctrStdin := make(chan string, 1)
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(ctrStdin),
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
	execStdin := make(chan string, 1)
	execStdout := make(chan string, 1)

	execResult := c.Exec(p,
		stream.WithStdinChannel(execStdin),
		stream.WithStdoutChannel(execStdout))

	select {
	case result := <-execResult:
		t.Fatalf("expected no result, got: %v", result.Err)
	default:
	}

	execStdin <- "Hello,\n"
	msg := <-execStdout
	if msg != "Hello,\n" {
		t.Fatalf("expected stdout message 'Hello,', got: %s", msg)
	}

	execStdin <- "world!\n"
	msg = <-execStdout
	if msg != "world!\n" {
		t.Fatalf("expected stdout message 'world!', got: %s", msg)
	}

	close(execStdin) // Signal end of input

	result := <-execResult
	AssertResult(t, result, "!", 0, nil)

	close(ctrStdin) // Allow container to close

	msg, ok := <-execStdout
	if ok {
		t.Fatalf("expected no message on stdout after closing stdin, got: %s", msg)
	}
	if msg != "" {
		t.Fatalf("expected empty message on stdout, got: %s", msg)
	}

	result = <-c.Done
	AssertResult(t, result, c.ID, 0, nil)
}
