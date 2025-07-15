package container_test

import (
	"errors"
	"testing"

	"manage2/internal"
	"manage2/internal/container"
	"manage2/internal/program"
	"manage2/internal/stream"
)

const commonImage = "alpine:latest"

func TestContainer(t *testing.T) {
	image := "some-image:latest"
	c := container.New(image)

	if c.Image != image {
		t.Fatalf("expected image '%s', got: %s", image, c.Image)
	}

	checkedRun(t, c, nil, container.ErrNoProgram)
}

func checkedRun(
	t *testing.T,
	c *container.Container,
	p *program.Program,
	expectErr error,
) {
	t.Helper()
	var err error

	if p == nil {
		err = c.Run()
	} else {
		err = c.Run(p)
	}

	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error '%v', got: %v", expectErr, err)
	}
	if c.ID == "" && expectErr == nil {
		t.Fatal("expected non-empty ID, got empty string")
	}
	if c.ID != "" && expectErr != nil {
		t.Fatalf("expected empty ID, got: %s", c.ID)
	}
}

func TestContainerWithProgram(t *testing.T) {
	c := container.New(commonImage)
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "exit 5"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)
	checkedDone(t, <-c.Done, c.ID, 5, nil)
}

func checkedDone(
	t *testing.T,
	r container.Result,
	expectedID string,
	expectedStatus int,
	expectedErr error,
) {
	t.Helper()
	if expectedID == "!" {
		if r.ID == "" {
			t.Errorf("expected non-empty ID, got: %s", r.ID)
		}
	} else {
		if r.ID != expectedID {
			t.Errorf("expected ID '%s', got: %s", expectedID, r.ID)
		}
	}
	if r.Status != expectedStatus {
		t.Errorf("expected exit code %d, got: %d", expectedStatus, r.Status)
	}
	if !errors.Is(r.Err, expectedErr) {
		t.Errorf("expected error '%v', got: %v", expectedErr, r.Err)
	}
}

func TestContainerWithStdout(t *testing.T) {
	ctrStdout := make(chan string)
	c := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	p, err := program.FromSystem("printf", program.WithStringArgs("|%s|\n", "Hello,", "world!"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	expectedOutput := "|Hello,|\n|world!|\n|--|\n"

	// ctrStdout channel is not buffered, so wait for the message
	// before waiting for the container to finish
	msg := <-ctrStdout
	if msg != expectedOutput {
		t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
	}

	checkedDone(t, <-c.Done, c.ID, 0, nil)

	// No further messages
	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages, got: %s", msg)
	}
}

func TestContainerWithStdoutAndStderr(t *testing.T) {
	ctrStdout := make(chan string)
	ctrStderr := make(chan string)
	c := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout),
		container.WithStderrChannel(ctrStderr))
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "printf '|%s|\n' Hello, world! | tee /dev/stderr"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	expectedOutput := "|Hello,|\n|world!|\n"

	for range 2 {
		select {
		case msg := <-ctrStdout:
			if msg != expectedOutput {
				t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
			}
		case msg := <-ctrStderr:
			if msg != expectedOutput {
				t.Fatalf("expected stderr message '%s', got: %s", expectedOutput, msg)
			}
		}
	}

	checkedDone(t, <-c.Done, c.ID, 0, nil)

	// No further messages
	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages on stdout, got: %s", msg)
	}
	msg, ok = <-ctrStderr
	if ok {
		t.Fatalf("expected no further messages on stderr, got: %s", msg)
	}
}

func TestContainerWithStdin(t *testing.T) {
	ctrStdin := make(chan string)
	ctrStdout := make(chan string)
	c := container.New(commonImage,
		container.WithStdinChannel(ctrStdin),
		container.WithStdoutChannel(ctrStdout))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && echo \"|$input|\""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Call Run() before sending input on ctrStdin because ctrStdin is
	// unbuffered. Writing to an unbuffered ctrStdin blocks until the stdin
	// streamer, which starts as part of Run(), reads from it (to write to the
	// container's stdin).
	checkedRun(t, c, p, nil)

	msg := "Hello, world!\n"
	ctrStdin <- msg
	close(ctrStdin)

	expectedMsg := "|Hello, world!|\n"
	msg = <-ctrStdout
	if msg != expectedMsg {
		t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
	}

	checkedDone(t, <-c.Done, c.ID, 0, nil)

	// No further messages
	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages, got: %s", msg)
	}
}

func TestContainerStdinOnly(t *testing.T) {
	ctrStdin := make(chan string)
	c := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	msg := "5\n"
	ctrStdin <- msg
	close(ctrStdin)

	checkedDone(t, <-c.Done, c.ID, 5, nil)
}

func TestBuildProgramCommandStrings(t *testing.T) {
	c := container.New(commonImage)
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
	file := dir.TouchFile("file.txt")

	c := container.New(commonImage)
	p, err := program.FromFile(script,
		program.WithStringArgs("--file"),
		program.WithFileArgs(file))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cmdTokens, fileMap := c.BuildProgramCmd(p)

	if len(fileMap) != 2 { // script & file
		t.Fatalf("expected 2 mounts, got: %d", len(fileMap))
	}

	expectedCmd := []string{fileMap[script], "--file", fileMap[file], "--"}
	if len(cmdTokens) != len(expectedCmd) {
		t.Fatalf("expected command tokens %v, got: %v", expectedCmd, cmdTokens)
	}

	for i, token := range cmdTokens {
		if token != expectedCmd[i] {
			t.Fatalf("expected command token '%s', got: '%s'", expectedCmd[i], token)
		}
	}
}

func TestContainerAutoMountsRunPrograms(t *testing.T) {
	dir := internal.NewTestDir(t)
	script := dir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"\n", 0755)
	expectedMsg := "Hello, world!\n"
	file := dir.WriteFile("file.txt", expectedMsg, 0644)

	ctrStdout := make(chan string)
	c := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	p, err := program.FromFile(script,
		program.WithFileArgs(file))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	msg := <-ctrStdout
	if msg != expectedMsg {
		t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
	}

	checkedDone(t, <-c.Done, c.ID, 0, nil)
}

// Assert we get strings exactly as written--we do not wait for e.g. newline
// characters before writing to the channel. This gives us more flexibility in
// how we interpret messages from the container, and we can choose to e.g. line
// buffer if we want to.
func TestContainerOutputAsIs(t *testing.T) {
	// TODO: Use an existing image with coreutils installed. Until then, get
	// coreutils and redirect apk output to an unused fd like stderr.
	script := internal.NewTestDir(t).WriteFile(
		"test.sh",
		`#!/bin/sh
apk add --no-cache coreutils >&2
stdbuf -o0 printf "Hello, "
stdbuf -o0 printf "world!\n"
`, 0755)
	ctrStdout := make(chan string)
	c := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	p, err := program.FromFile(script)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	msg := <-ctrStdout
	expectedMsg := "Hello, "
	if msg != expectedMsg {
		t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
	}

	msg = <-ctrStdout
	expectedMsg = "world!\n"
	if msg != expectedMsg {
		t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
	}

	checkedDone(t, <-c.Done, c.ID, 0, nil)

	// No further messages
	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages, got: %s", msg)
	}
}

// Assert we can control when a container exits by running a command that waits
// for input on stdin like `cat`. This allows us to explicitly control when the
// container closes by closing the stdin channel. This is useful to run multiple
// arbitrary commands with Exec() before closing the container.
func TestIdleContainer(t *testing.T) {
	ctrStdin := make(chan string)
	c := container.New(commonImage, container.WithStdinChannel(ctrStdin))
	p, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	close(ctrStdin) // Allow container to close
	checkedDone(t, <-c.Done, c.ID, 0, nil)
}

func TestExec(t *testing.T) {
	ctrStdin := make(chan string)
	c := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, p, nil)

	execStdin := make(chan string)
	execResult := c.Exec(p, stream.WithStdinChannel(execStdin))

	select {
	case result := <-execResult:
		t.Fatalf("expected no result, got: %+v", result)
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
	c := container.New(commonImage)
	p, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	checkedDone(t, <-c.Exec(p), "", -1, container.ErrNotRunning)
}

func TestExecNoPrograms(t *testing.T) {
	c := container.New(commonImage)
	checkedDone(t, <-c.Exec(nil), "", -1, container.ErrNoProgram)
}

func TestExecMissingMounts(t *testing.T) {
	ctrStdin := make(chan string)
	c := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, wait, nil)

	path := internal.NewTestDir(t).TouchFile("file.txt")
	p, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Must wait for exec program to finish before closing the container!
	checkedDone(t, <-c.Exec(p), "", -1, container.ErrNotMounted)

	close(ctrStdin) // Allow container to close

	checkedDone(t, <-c.Done, c.ID, 0, nil)
}

func TestExecMounts(t *testing.T) {
	ctrStdin := make(chan string)
	path := internal.NewTestDir(t).TouchFile("file.txt")
	c := container.New(commonImage,
		container.WithStdinChannel(ctrStdin),
		container.WithMounts(path))

	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, c, wait, nil)

	p, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Must wait for exec program to finish before closing the container!
	checkedDone(t, <-c.Exec(p), "!", 0, nil)

	close(ctrStdin) // Allow container to close

	checkedDone(t, <-c.Done, c.ID, 0, nil)
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
	ctrStdin := make(chan string)
	c := container.New(commonImage,
		container.WithStdinChannel(ctrStdin),
		container.WithMounts(path))

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
	execStdin := make(chan string)
	execStdout := make(chan string)

	execResult := c.Exec(p,
		stream.WithStdinChannel(execStdin),
		stream.WithStdoutChannel(execStdout))

	select {
	case result := <-execResult:
		t.Fatalf("expected no result, got: %+v", result)
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

	checkedDone(t, <-execResult, "!", 0, nil)

	close(ctrStdin) // Allow container to close

	msg, ok := <-execStdout
	if ok {
		t.Fatalf("expected no message on stdout after closing stdin, got: %s", msg)
	}
	if msg != "" {
		t.Fatalf("expected empty message on stdout, got: %s", msg)
	}

	checkedDone(t, <-c.Done, c.ID, 0, nil)
}
