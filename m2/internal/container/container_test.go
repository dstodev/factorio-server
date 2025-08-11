package container_test

import (
	"errors"
	"fmt"
	"testing"

	"manage2/internal/container"
	"manage2/internal/container/stream"
	"manage2/internal/program"
	"manage2/internal/test"
)

const commonImage = "m2test:latest" // See Makefile target: test-image

// #region Test Container IO
func TestContainer(t *testing.T) {
	t.Parallel()
	image := "some-image:latest"
	ctr := container.New(image)

	if ctr.Image != image {
		t.Fatalf("expected image '%s', received: %s", image, ctr.Image)
	}

	checkedRun(t, ctr, container.ErrNoProgram)
}

func checkedRun(
	t *testing.T,
	ctr *container.Container,
	expectErr error,
	programs ...*program.Program,
) (
	ctrDone <-chan container.Result,
) {
	t.Helper()

	ctrDone, err := ctr.Run(programs...)

	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error '%v', received: %v", expectErr, err)
	}
	if ctr.ID == "" && expectErr == nil {
		t.Fatal("expected non-empty ID, received empty string")
	}
	if ctr.ID != "" && expectErr != nil {
		t.Fatalf("expected empty ID, received: %s", ctr.ID)
	}
	return
}

func TestContainerWithProgram(t *testing.T) {
	t.Parallel()
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "exit 5"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)
	checkResult(t, <-ctrDone, ctr.ID, 5, nil)
}

func TestContainerOnlyRunsOnce(t *testing.T) {
	t.Parallel()
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "exit 5"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)
	checkResult(t, <-ctrDone, ctr.ID, 5, nil)

	ctrDone, err = ctr.Run(pgm)
	if !errors.Is(err, container.ErrRunning) {
		t.Fatalf("expected error '%v', received: %v", container.ErrRunning, err)
	}
	if ctrDone != nil {
		t.Fatalf("expected nil ctrDone channel, received: %v", ctrDone)
	}
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
			t.Errorf("expected non-empty ID, received: %s", result.ID)
		}
	} else {
		if result.ID != expectedID {
			t.Errorf("expected ID '%s', received: %s", expectedID, result.ID)
		}
	}
	if result.Status != expectedStatus {
		t.Errorf("expected exit code %d, received: %d", expectedStatus, result.Status)
	}
	if !errors.Is(result.Err, expectedErr) {
		t.Errorf("expected error '%v', received: %v", expectedErr, result.Err)
	}
}

func TestContainerWithStdout(t *testing.T) {
	t.Parallel()
	catStdout := stream.NewCat()
	ctr := container.New(commonImage,
		container.WithStdoutChannel(catStdout.Stdin))
	pgm, err := program.FromSystem("printf", program.WithStringArgs("|%s|\n", "Hello,", "world!"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)
	checkResult(t, <-ctrDone, ctr.ID, 0, nil)

	expected := "|Hello,|\n|world!|\n|--|\n"
	received := catStdout.String()
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}
}

func TestContainerWithStdoutAndStderr(t *testing.T) {
	t.Parallel()
	catStdout := stream.NewCat()
	catStderr := stream.NewCat()
	ctr := container.New(commonImage,
		container.WithStdoutChannel(catStdout.Stdin),
		container.WithStderrChannel(catStderr.Stdin))
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "printf '|%s|\n' Hello, world! | tee /dev/stderr"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)
	checkResult(t, <-ctrDone, ctr.ID, 0, nil)

	expected := "|Hello,|\n|world!|\n"
	received := catStdout.String()
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}
	received = catStderr.String()
	if expected != received {
		t.Fatalf("expected stderr message '%s', received: %s", expected, received)
	}
}

func TestContainerWithStdin(t *testing.T) {
	t.Parallel()
	ctrStdin := make(chan string)
	ctrStdout := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin),
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "read input && echo \"|$input|\""))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	// Call Run() before sending input on ctrStdin because ctrStdin is not
	// buffered. Writing to an unbuffered channel blocks until read, done here
	// by the stdin streamer, which starts as part of Run().
	ctrDone := checkedRun(t, ctr, nil, pgm)

	// `read` waits for newline, so we must send a newline after each message.
	ctrStdin <- "Hello, world!\n"
	close(ctrStdin)

	expected := "|Hello, world!|\n"
	received := <-ctrStdout
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}

	checkResult(t, <-ctrDone, ctr.ID, 0, nil)

	select {
	case msg, ok := <-ctrStdout:
		if ok {
			t.Fatalf("expected no further messages, received: %s", msg)
		}
	default:
		t.Fatalf("expected channel to close")
	}
}

func TestContainerStdinOnly(t *testing.T) {
	t.Parallel()
	ctrStdin := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	pgm, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)

	msg := "5\n"
	ctrStdin <- msg
	close(ctrStdin)

	checkResult(t, <-ctrDone, ctr.ID, 5, nil)
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
	t.Parallel()
	workdir := test.Workdir(t)
	script := workdir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"", 0o755)
	expected := "Hello, world!\n"
	file := workdir.WriteFile("file.txt", expected, 0o644)

	ctrStdout := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromFile(script,
		program.WithFileArgs(file))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)

	received := <-ctrStdout
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}

	checkResult(t, <-ctrDone, ctr.ID, 0, nil)
}

// Assert we get strings exactly as written--we do not wait for e.g. newline
// characters before writing to the channel. This gives us more flexibility in
// how we interpret messages from the container, and we can choose to e.g. line
// buffer if we want to.
func TestContainerOutputAsIs(t *testing.T) {
	t.Parallel()
	script := test.Workdir(t).WriteFile(
		"test.sh",
		`#!/bin/sh
stdbuf -o0 printf "Hello, "
stdbuf -o0 printf "world!\n"
`, 0o755)
	catStdout := stream.NewCat()
	ctr := container.New(commonImage,
		container.WithStdoutChannel(catStdout.Stdin))
	pgm, err := program.FromFile(script)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)
	checkResult(t, <-ctrDone, ctr.ID, 0, nil)

	expected := "Hello, world!\n"
	received := catStdout.String()
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}
}

func TestProgramChain(t *testing.T) {
	t.Parallel()
	script := test.Workdir(t).WriteFile("script.sh", `#!/bin/sh
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
`, 0o755)

	catStdout := stream.NewCat()
	ctr := container.New(commonImage,
		container.WithStdoutChannel(catStdout.Stdin))

	pgm1, err := program.FromFile(script, program.WithStringArgs("-xy", "--optY"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	pgm2, err := program.FromFile(script, program.WithStringArgs("-yx", "--optX"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm1, pgm2)
	checkResult(t, <-ctrDone, ctr.ID, 0, nil)

	expected := fmt.Sprintf(`
Args: -x -y --optY -- %s -yx --optX --
Option X is set
Option Y is set

Args: -y -x --optX --
Option X is set
Option Y is set
`, ctr.HostToGuestMap[script])

	received := catStdout.String()

	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}
}

// #endregion

// #region Test Container utils

func TestBuildProgramCommandStrings(t *testing.T) {
	t.Parallel()
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "echo Hello"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	cmdTokens, mounts := container.BuildProgramCmd(ctr.HostToGuestMap, pgm)

	expectedCmd := []string{"/bin/sh", "-c", "echo Hello", "--"}
	if len(cmdTokens) != len(expectedCmd) {
		t.Fatalf("expected command tokens %v, received: %v", expectedCmd, cmdTokens)
	}
	for i, token := range cmdTokens {
		if token != expectedCmd[i] {
			t.Fatalf("expected command token '%s', received: '%s'", expectedCmd[i],
				token)
		}
	}
	if len(mounts) != 0 {
		t.Fatalf("expected no mapping, received: %v", mounts)
	}
}

func TestBuildProgramCommandWithFileArgs(t *testing.T) {
	t.Parallel()
	workdir := test.Workdir(t)
	script := workdir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"", 0755)
	file := workdir.TouchFile("file.txt")

	ctr := container.New(commonImage)
	pgm, err := program.FromFile(script,
		program.WithStringArgs("--file"),
		program.WithFileArgs(file),
		program.WithFileArgs(script))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	cmdTokens, fileMap := container.BuildProgramCmd(ctr.HostToGuestMap, pgm)

	if len(fileMap) != 2 { // script & file
		t.Fatalf("expected 2 mapping, received: %d", len(fileMap))
	}

	expectedCmd := []string{fileMap[script], "--file", fileMap[file], fileMap[script], "--"}
	if len(cmdTokens) != len(expectedCmd) {
		t.Fatalf("expected command tokens %v, received: %v", expectedCmd, cmdTokens)
	}

	for i, token := range cmdTokens {
		if token != expectedCmd[i] {
			t.Fatalf("expected command token '%s', received: '%s'", expectedCmd[i], token)
		}
	}
}

func TestBuildProgramNewMappings(t *testing.T) {
	t.Parallel()
	workdir := test.Workdir(t)
	script := workdir.WriteFile("test.sh", "#!/bin/sh\ncat \"$@\"", 0755)
	file := workdir.TouchFile("file.txt")

	ctr := container.New(commonImage)
	pgm, err := program.FromFile(script,
		program.WithStringArgs("--file"),
		program.WithFileArgs(file),
		program.WithFileArgs(script))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	_, fileMap := container.BuildProgramCmd(ctr.HostToGuestMap, pgm)

	if len(fileMap) != 2 { // script & file
		t.Fatalf("expected 2 mapping, received: %d", len(fileMap))
	}

	// Test idempotence
	_, fileMap = container.BuildProgramCmd(ctr.HostToGuestMap, pgm)
	if len(fileMap) != 2 {
		t.Fatalf("expected 2 mapping, received: %d", len(fileMap))
	}

	// Test recording a mapping to the container's HostToGuestMap
	// BuildProgramCmd() should not return it again, since it is no longer new.
	ctr.HostToGuestMap[script] = fileMap[script]

	_, fileMap = container.BuildProgramCmd(ctr.HostToGuestMap, pgm)
	if len(fileMap) != 1 {
		t.Fatalf("expected 1 mapping, received: %d", len(fileMap))
	}
	if _, ok := fileMap[script]; ok {
		t.Fatalf("expected no new mapping for script, received: %s", fileMap[script])
	}
}

// Assert we can control when a container exits by running a command that waits
// for input on stdin like `cat`. This allows us to explicitly control when the
// container closes by closing the stdin channel. This is useful to run multiple
// arbitrary commands with Exec() before closing the container.
func TestIdleContainer(t *testing.T) {
	t.Parallel()
	ctr, ctrClose := container.Idle(commonImage)
	checkResult(t, ctrClose(), ctr.ID, 0, nil)
}

// #endregion

// #region Test Container.Exec()

func TestExec(t *testing.T) {
	t.Parallel()
	ctrStdin := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin))
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "read input && exit $input"))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)

	execStdin := make(chan string)
	execDone, err := ctr.Exec(pgm, stream.WithStdinChannel(execStdin))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	select {
	case result := <-execDone:
		t.Fatalf("expected no result, received: %+v", result)
	default:
	}

	execStdin <- "4\n"
	close(execStdin)

	checkResult(t, <-execDone, "!", 4, nil)

	ctrStdin <- "5\n"
	close(ctrStdin)

	checkResult(t, <-ctrDone, ctr.ID, 5, nil)
}

func TestExecNotRunning(t *testing.T) {
	t.Parallel()
	ctr := container.New(commonImage)
	pgm, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	execDone, err := ctr.Exec(pgm)
	if !errors.Is(err, container.ErrNotRunning) {
		t.Fatalf("expected error '%v', received: %v", container.ErrNotRunning, err)
	}
	if execDone != nil {
		t.Fatalf("expected nil execDone channel, received: %v", execDone)
	}
}

func TestExecNoPrograms(t *testing.T) {
	t.Parallel()
	ctr := container.New(commonImage)
	execDone, err := ctr.Exec(nil)
	if !errors.Is(err, container.ErrNoProgram) {
		t.Fatalf("expected error '%v', received: %v", container.ErrNoProgram, err)
	}
	if execDone != nil {
		t.Fatalf("expected nil execDone channel, received: %v", execDone)
	}
}

func TestExecMissingMounts(t *testing.T) {
	t.Parallel()
	ctr, ctrClose := container.Idle(commonImage)

	path := test.Workdir(t).TouchFile("file.txt")
	pgm, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	execDone, err := ctr.Exec(pgm)
	if !errors.Is(err, container.ErrNotMounted) {
		t.Fatalf("expected error '%v', received: %v", container.ErrNotMounted, err)
	}

	if execDone != nil {
		t.Fatalf("expected nil execDone channel, received: %v", execDone)
	}

	checkResult(t, ctrClose(), ctr.ID, 0, nil)
}

func TestExecMounts(t *testing.T) {
	t.Parallel()
	path := test.Workdir(t).TouchFile("file.txt")
	ctr, ctrClose := container.Idle(commonImage, container.WithMounts(path))

	pgm, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	execDone, err := ctr.Exec(pgm)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	// Wait for exec program to finish before closing the container!
	checkResult(t, <-execDone, "!", 0, nil)
	checkResult(t, ctrClose(), ctr.ID, 0, nil)
}

func TestExecStdinStdout(t *testing.T) {
	t.Parallel()
	script := test.Workdir(t).WriteFile("script.sh", `#!/bin/sh
while :; do
	read line
	if [ -z "$line" ]; then
		break
	fi
	echo "$line"
done
`, 0o744)
	ctr, ctrClose := container.Idle(commonImage, container.WithMounts(script))

	pgm, err := program.FromFile(script)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	execStdin := make(chan string)
	execStdout := make(chan string)

	execDone, err := ctr.Exec(pgm,
		stream.WithStdinChannel(execStdin),
		stream.WithStdoutChannel(execStdout))

	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	execStdin <- "Hello,\n"
	msg := <-execStdout
	if msg != "Hello,\n" {
		t.Fatalf("expected stdout message 'Hello,', received: %s", msg)
	}

	execStdin <- "world!\n"
	msg = <-execStdout
	if msg != "world!\n" {
		t.Fatalf("expected stdout message 'world!', received: %s", msg)
	}

	close(execStdin) // Signal end of input

	checkResult(t, <-execDone, "!", 0, nil)
	checkResult(t, ctrClose(), ctr.ID, 0, nil)

	select {
	case msg, ok := <-execStdout:
		if ok {
			t.Fatalf("expected no further messages, received: %s", msg)
		}
	default:
		t.Fatalf("expected channel to close")
	}
}

// #endregion

// #region Test Container.Find()

func TestFindContainerByID(t *testing.T) {
	t.Parallel()
	ctr, ctrClose := container.Idle(commonImage)
	foundCtr, err := container.Find(ctr.ID)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	if foundCtr == nil {
		t.Fatal("expected to find container, received nil")
	}
	if foundCtr.ID != ctr.ID {
		t.Fatalf("expected found container ID '%s', received: %s", ctr.ID, foundCtr.ID)
	}
	checkResult(t, ctrClose(), ctr.ID, 0, nil)
}

func TestFindContainerByName(t *testing.T) {
	t.Parallel()
	ctrName := "test-container"

	foundCtr, err := container.Find(ctrName)

	if !errors.Is(err, container.ErrNotFound) {
		t.Fatalf("expected error '%v', received: %v", container.ErrNotFound, err)
	}
	if foundCtr != nil {
		t.Fatalf("expected nil container, received: %v", foundCtr)
	}
	_, ctrClose := container.Idle(commonImage,
		container.WithName(ctrName))

	foundCtr, err = container.Find(ctrName)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	if foundCtr == nil {
		t.Fatal("expected to find container, received nil")
	}
	if foundCtr.ID == "" {
		t.Fatal("expected non-empty ID, received empty string")
	}
	if foundCtr.Name != ctrName {
		t.Fatalf("expected container name '%s', received: %s", ctrName, foundCtr.Name)
	}

	checkResult(t, ctrClose(), foundCtr.ID, 0, nil)
}

func TestFindContainerNotFound(t *testing.T) {
	t.Parallel()
	ctr, err := container.Find("__does-not-exist")
	if !errors.Is(err, container.ErrNotFound) {
		t.Fatalf("expected error '%v', received: %v", container.ErrNotFound, err)
	}
	if ctr != nil {
		t.Fatalf("expected nil container, received: %v", ctr)
	}
}

func TestFindContainerExec(t *testing.T) {
	t.Parallel()
	ctrStdin := make(chan string)
	ctrStdout := make(chan string)
	ctr := container.New(commonImage,
		container.WithStdinChannel(ctrStdin),
		container.WithStdoutChannel(ctrStdout))
	pgm, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "read input && echo \"|$input|\""))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	ctrDone := checkedRun(t, ctr, nil, pgm)

	ctrStdin <- "Hello, "

	foundCtr, err := container.Find(ctr.ID)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	if foundCtr == nil {
		t.Fatal("expected to find container, received nil")
	}
	if foundCtr.ID != ctr.ID {
		t.Fatalf("expected found container ID '%s', received: %s", ctr.ID, foundCtr.ID)
	}

	execStdin := make(chan string)
	execStdout := make(chan string)
	execDone, err := foundCtr.Exec(pgm,
		stream.WithStdinChannel(execStdin),
		stream.WithStdoutChannel(execStdout))
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}

	execStdin <- "wow!\n"
	close(execStdin)

	expected := "|wow!|\n"
	received := <-execStdout
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}
	checkResult(t, <-execDone, "!", 0, nil)

	ctrStdin <- "world!\n"

	received = <-ctrStdout
	expected = "|Hello, world!|\n"
	if expected != received {
		t.Fatalf("expected stdout message '%s', received: %s", expected, received)
	}

	close(ctrStdin)
	checkResult(t, <-ctrDone, ctr.ID, 0, nil)
}

func TestFindContainerRunNotAllowed(t *testing.T) {
	t.Parallel()
	ctr, ctrClose := container.Idle(commonImage)

	foundCtr, err := container.Find(ctr.ID)
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	if foundCtr == nil {
		t.Fatal("expected to find container, received nil")
	}
	if foundCtr.ID != ctr.ID {
		t.Fatalf("expected found container ID '%s', received: %s", ctr.ID, foundCtr.ID)
	}

	ctrDone, err := foundCtr.Run()
	if !errors.Is(err, container.ErrRunning) {
		t.Fatalf("expected error '%v', received: %v", container.ErrRunning, err)
	}
	if ctrDone != nil {
		t.Fatalf("expected nil ctrDone channel, received: %v", ctrDone)
	}

	checkResult(t, ctrClose(), ctr.ID, 0, nil)
}
