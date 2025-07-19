package container_test

import (
	"errors"
	"fmt"
	"testing"

	"manage2/internal"
	"manage2/internal/container"
	"manage2/internal/program"
	"manage2/internal/stream"
)

const commonImage = "m2test:latest" // See Makefile target: test-image

// #region Test Container IO
func TestContainer(t *testing.T) {
	image := "some-image:latest"
	ctr := container.New(image)

	if ctr.Image != image {
		t.Fatalf("expected image '%s', got: %s", image, ctr.Image)
	}

	checkedRun(t, ctr, container.ErrNoProgram)
}

func checkedRun(
	t *testing.T,
	ctr *container.Container,
	expectErr error,
	programs ...*program.Program,
) {
	t.Helper()

	err := ctr.Run(programs...)

	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error '%v', got: %v", expectErr, err)
	}
	if ctr.ID == "" && expectErr == nil {
		t.Fatal("expected non-empty ID, got empty string")
	}
	if ctr.ID != "" && expectErr != nil {
		t.Fatalf("expected empty ID, got: %s", ctr.ID)
	}
}

func TestContainerWithProgram(t *testing.T) {
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "exit 5"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)
	checkResult(t, <-ctr.Done, ctr.ID, 5, nil)
}

func checkResult(
	t *testing.T,
	result container.Result,
	expectedID string,
	expectedStatus int,
	expectedErr error,
) {
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
	ctrStdout := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromSystem("printf", program.WithStringArgs("|%s|\n", "Hello,", "world!"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)

	expectedOutput := "|Hello,|\n|world!|\n|--|\n"

	// ctrStdout channel is not buffered, so wait for the message
	// before waiting for the container to finish
	msg := <-ctrStdout
	if msg != expectedOutput {
		t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, msg)
	}

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)

	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages, got: %s", msg)
	}
}

func TestContainerWithStdoutAndStderr(t *testing.T) {
	ctrStdout := make(chan string)
	ctrStderr := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout),
		container.WithStderrChannel(ctrStderr))
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "printf '|%s|\n' Hello, world! | tee /dev/stderr"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)

	expectedOutput := "|Hello,|\n|world!|\n"

	for range 2 { // expect two messages, one on stdout and one on stderr
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

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)

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
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin),
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && echo \"|$input|\""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Call Run() before sending input on ctrStdin because ctrStdin is not
	// buffered. Writing to an unbuffered channel blocks until read, done here
	// by the stdin streamer, which starts as part of Run().
	checkedRun(t, ctr, nil, pgm)

	// `read` waits for newline, so we must send a newline after each message.
	msg := "Hello, world!\n"
	ctrStdin <- msg
	close(ctrStdin)

	expectedMsg := "|Hello, world!|\n"
	msg = <-ctrStdout
	if msg != expectedMsg {
		t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
	}

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)

	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages, got: %s", msg)
	}
}

func TestContainerStdinOnly(t *testing.T) {
	ctrStdin := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	pgm, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)

	msg := "5\n"
	ctrStdin <- msg
	close(ctrStdin)

	checkResult(t, <-ctr.Done, ctr.ID, 5, nil)
}

// When a file is specified using program.FromFile() or program.WithFileArgs(),
// that file is automatically mounted into the container. When referenced as an
// argument using program.WithFileArgs(), the path is translated before passing
// it to the program as an argument:
//
//	pgm, err := program.FromFile(script, program.WithFileArgs(file))
//
//	/host/path/to/script.sh -> /tmp/m2/3015c825-186c-4d02-9198-f21873ca544d/test.sh
//	/host/path/to/file.txt -> /tmp/m2/46183341-9a95-49c2-9e0c-0d654dcba00f/file.txt
//
//	Container command line: [
//	    "/tmp/m2/b440f488-c7fc-490b-bd6f-031887da09a2/test.sh",
//	    "/tmp/m2/c31e44f9-af2b-4030-9b8b-786d811dd4e0/file.txt",
//	    "--"]
func TestContainerAutoMountsRunPrograms(t *testing.T) {
	dir := internal.NewTestDir(t)
	script := dir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"\n", 0755)
	expectedMsg := "Hello, world!\n"
	file := dir.WriteFile("file.txt", expectedMsg, 0644)

	ctrStdout := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromFile(script,
		program.WithFileArgs(file))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)

	msg := <-ctrStdout
	if msg != expectedMsg {
		t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
	}

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)
}

// Assert we get strings exactly as written--we do not wait for e.g. newline
// characters before writing to the channel. This gives us more flexibility in
// how we interpret messages from the container, and we can choose to e.g. line
// buffer if we want to.
func TestContainerOutputAsIs(t *testing.T) {
	script := internal.NewTestDir(t).WriteFile(
		"test.sh",
		`#!/bin/sh
stdbuf -o0 printf "Hello, "
stdbuf -o0 printf "world!\n"
`, 0755)
	ctrStdout := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromFile(script)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)

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

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)

	msg, ok := <-ctrStdout
	if ok {
		t.Fatalf("expected no further messages, got: %s", msg)
	}
}

func TestProgramChain(t *testing.T) {
	script := internal.NewTestDir(t).WriteFile("script.sh", `#!/bin/sh
canonical=$(getopt --name "$(basename "$0")" \
	--options xy \
	--longoptions optX,optY \
	-- "$@")

eval set -- "$canonical"

echo
echo "Args: $@"

while :; do
	arg="$1"
	shift

	case "$arg" in
	-x | --optX)
		optX=true
		;;
	-y | --optY)
		optY=true
		;;
	--)
		break
		;;
	esac
done

optX=${optX-false}
optY=${optY-false}

if $optX; then
	echo "Option X is set"
fi
if $optY; then
	echo "Option Y is set"
fi

"$@"
`, 0755)

	catStdout := stream.NewStringCat()
	ctr := container.New(commonImage,
		container.WithStdoutChannel(catStdout.Stdin))

	pgm1, err := program.FromFile(script, program.WithStringArgs("-xy", "--optY"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	pgm2, err := program.FromFile(script, program.WithStringArgs("-yx", "--optX"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm1, pgm2)
	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)

	expectedOutput := fmt.Sprintf(`
Args: -x -y --optY -- %s -yx --optX --
Option X is set
Option Y is set

Args: -y -x --optX --
Option X is set
Option Y is set
`, ctr.HostToGuestMap[script])

	stdout := catStdout.Print()

	if stdout != expectedOutput {
		t.Fatalf("expected stdout message '%s', got: %s", expectedOutput, stdout)
	}
}

// #endregion

// #region Test Container utils

func TestBuildProgramCommandStrings(t *testing.T) {
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "echo Hello"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cmdTokens, mounts := ctr.BuildProgramCmd(pgm)

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
		t.Fatalf("expected no mapping, got: %v", mounts)
	}
}

func TestBuildProgramCommandWithFileArgs(t *testing.T) {
	dir := internal.NewTestDir(t)
	script := dir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"\n", 0755)
	file := dir.TouchFile("file.txt")

	ctr := container.New(commonImage)
	pgm, err := program.FromFile(script,
		program.WithStringArgs("--file"),
		program.WithFileArgs(file),
		program.WithFileArgs(script))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cmdTokens, fileMap := ctr.BuildProgramCmd(pgm)

	if len(fileMap) != 2 { // script & file
		t.Fatalf("expected 2 mapping, got: %d", len(fileMap))
	}

	expectedCmd := []string{fileMap[script], "--file", fileMap[file], fileMap[script], "--"}
	if len(cmdTokens) != len(expectedCmd) {
		t.Fatalf("expected command tokens %v, got: %v", expectedCmd, cmdTokens)
	}

	for i, token := range cmdTokens {
		if token != expectedCmd[i] {
			t.Fatalf("expected command token '%s', got: '%s'", expectedCmd[i], token)
		}
	}
}

func TestBuildProgramNewMappings(t *testing.T) {
	dir := internal.NewTestDir(t)
	script := dir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"\n", 0755)
	file := dir.TouchFile("file.txt")

	ctr := container.New(commonImage)
	pgm, err := program.FromFile(script,
		program.WithStringArgs("--file"),
		program.WithFileArgs(file),
		program.WithFileArgs(script))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, fileMap := ctr.BuildProgramCmd(pgm)

	if len(fileMap) != 2 { // script & file
		t.Fatalf("expected 2 mapping, got: %d", len(fileMap))
	}

	// Test idempotence
	_, fileMap = ctr.BuildProgramCmd(pgm)
	if len(fileMap) != 2 {
		t.Fatalf("expected 2 mapping, got: %d", len(fileMap))
	}

	// Test recording a mapping to the container's HostToGuestMap
	// BuildProgramCmd() should not return it again, since it is no longer new.
	ctr.HostToGuestMap[script] = fileMap[script]

	_, fileMap = ctr.BuildProgramCmd(pgm)
	if len(fileMap) != 1 {
		t.Fatalf("expected 1 mapping, got: %d", len(fileMap))
	}
	if _, ok := fileMap[script]; ok {
		t.Fatalf("expected no new mapping for script, got: %s", fileMap[script])
	}
}

// Assert we can control when a container exits by running a command that waits
// for input on stdin like `cat`. This allows us to explicitly control when the
// container closes by closing the stdin channel. This is useful to run multiple
// arbitrary commands with Exec() before closing the container.
func TestIdleContainer(t *testing.T) {
	ctr, ctrClose := container.Idle(commonImage)
	ctrClose() // Test times out without calling this function
	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)
}

// #endregion

// #region Test Container.Exec

func TestExec(t *testing.T) {
	ctrStdin := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	checkedRun(t, ctr, nil, pgm)

	execStdin := make(chan string)
	execDone, err := ctr.Exec(pgm, stream.WithStdinChannel(execStdin))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	select {
	case result := <-execDone:
		t.Fatalf("expected no result, got: %+v", result)
	default:
	}

	execStdin <- "4\n"
	close(execStdin)

	checkResult(t, <-execDone, "!", 4, nil)

	ctrStdin <- "5\n"
	close(ctrStdin)

	checkResult(t, <-ctr.Done, ctr.ID, 5, nil)
}

func TestExecNotRunning(t *testing.T) {
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	execDone, err := ctr.Exec(pgm)
	if err != container.ErrNotRunning {
		t.Fatalf("expected error '%v', got: %v", container.ErrNotRunning, err)
	}
	if execDone != nil {
		t.Fatalf("expected nil execDone channel, got: %v", execDone)
	}
}

func TestExecNoPrograms(t *testing.T) {
	ctr := container.New(commonImage)
	execDone, err := ctr.Exec(nil)
	if err != container.ErrNoProgram {
		t.Fatalf("expected error '%v', got: %v", container.ErrNoProgram, err)
	}
	if execDone != nil {
		t.Fatalf("expected nil execDone channel, got: %v", execDone)
	}
}

func TestExecMissingMounts(t *testing.T) {
	ctr, ctrClose := container.Idle(commonImage)

	path := internal.NewTestDir(t).TouchFile("file.txt")
	pgm, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	execDone, err := ctr.Exec(pgm)
	if err != container.ErrNotMounted {
		t.Fatalf("expected error '%v', got: %v", container.ErrNotMounted, err)
	}

	if execDone != nil {
		t.Fatalf("expected nil execDone channel, got: %v", execDone)
	}

	ctrClose()

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)
}

func TestExecMounts(t *testing.T) {
	path := internal.NewTestDir(t).TouchFile("file.txt")
	ctr, ctrClose := container.Idle(commonImage, container.WithMounts(path))

	pgm, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	execDone, err := ctr.Exec(pgm)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Wait for exec program to finish before closing the container!
	checkResult(t, <-execDone, "!", 0, nil)

	ctrClose()

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)
}

func TestExecStdinWrite(t *testing.T) {
	script := internal.NewTestDir(t).WriteFile("script.sh", `#!/bin/sh
while :; do
	read line
	if [ -z "$line" ]; then
		break
	fi
	echo "$line"
done
`, 0744)
	ctr, ctrClose := container.Idle(commonImage, container.WithMounts(script))

	pgm, err := program.FromFile(script)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	execStdin := make(chan string)
	execStdout := make(chan string)

	execDone, err := ctr.Exec(pgm,
		stream.WithStdinChannel(execStdin),
		stream.WithStdoutChannel(execStdout))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
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

	checkResult(t, <-execDone, "!", 0, nil)

	ctrClose()

	msg, ok := <-execStdout
	if ok {
		t.Fatalf("expected no message on stdout after closing stdin, got: %s", msg)
	}
	if msg != "" {
		t.Fatalf("expected empty message on stdout, got: %s", msg)
	}

	checkResult(t, <-ctr.Done, ctr.ID, 0, nil)
}

// #endregion
