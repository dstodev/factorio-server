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
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "exit 5"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	c := internal.NewContainer(commonImage)

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 5 {
		t.Fatalf("expected exit code 5, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}
}

func TestContainerWithStdout(t *testing.T) {
	p, err := program.FromSystem("printf", program.WithStringArgs("|%s|\n", "Hello,", "world!"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdoutChan := make(chan string, 1)

	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(stdoutChan),
	)

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	expectedOutput := "|Hello,|\n|world!|\n|--|\n"

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}

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
	p, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "printf '|%s|\n' Hello, world! | tee /dev/stderr"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdoutChan := make(chan string, 1)
	stderrChan := make(chan string, 1)

	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(stdoutChan),
		internal.WithStderrChannel(stderrChan),
	)

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	expectedOutput := "|Hello,|\n|world!|\n"

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}

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
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && echo \"|$input|\""))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdinChan := make(chan string, 1)
	stdoutChan := make(chan string, 1)

	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stdinChan),
		internal.WithStdoutChannel(stdoutChan),
	)

	msg := "Hello, world!\n"
	expectedMsg := "|Hello, world!|\n"

	stdinChan <- msg
	close(stdinChan)

	if err := c.Run(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if c.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}

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
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdinChan := make(chan string, 1)

	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stdinChan),
	)

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
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 5 {
		t.Fatalf("expected exit code 5, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}
}

func TestIdleContainer(t *testing.T) {
	stdinChan := make(chan string, 1)
	c := internal.NewContainer(commonImage, internal.WithStdinChannel(stdinChan))
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
	close(stdinChan) // Test times out without closing stdin; container is idle
	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}
}

func TestBuildProgramCommandStrings(t *testing.T) {
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "echo Hello"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	c := internal.NewContainer(commonImage)
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

	p, err := program.FromFile(script,
		program.WithFileArgs(file),
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdoutChan := make(chan string, 1)

	c := internal.NewContainer(commonImage,
		internal.WithStdoutChannel(stdoutChan))

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
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}

	select {
	case msg := <-stdoutChan:
		if msg != expectedMsg {
			t.Fatalf("expected stdout message '%s', got: %s", expectedMsg, msg)
		}
	default:
		t.Fatal("expected a message on stdout")
	}
}

func TestExec(t *testing.T) {
	ctrChan := make(chan string, 1)
	c := internal.NewContainer(commonImage, internal.WithStdinChannel(ctrChan))
	p, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && exit $input"))
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

	ctrChan <- "5\n"
	close(ctrChan)

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
	if !errors.Is(result.Err, internal.ErrNotRunning) {
		t.Fatalf("expected error '%v', got: %v", internal.ErrNotRunning, result.Err)
	}
	if result.Status != -1 {
		t.Fatalf("expected exit code -1, got: %d", result.Status)
	}
	if result.ID != "" {
		t.Fatal("expected empty ID, got non-empty string")
	}
}

func TestExecNoPrograms(t *testing.T) {
	c := internal.NewContainer(commonImage)

	resultChan := c.Exec(nil)

	result := <-resultChan
	if !errors.Is(result.Err, internal.ErrNoProgram) {
		t.Fatalf("expected error '%v', got: %v", internal.ErrNoProgram, result.Err)
	}
	if result.Status != -1 {
		t.Fatalf("expected exit code -1, got: %d", result.Status)
	}
	if result.ID != "" {
		t.Fatal("expected empty ID, got non-empty string")
	}
}

func TestExecMissingMounts(t *testing.T) {
	stdinChan := make(chan string, 1)
	c := internal.NewContainer(commonImage, internal.WithStdinChannel(stdinChan))
	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := c.Run(wait); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	path := internal.NewTestDir(t).TouchFile("file.txt")
	p, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	resultChan := c.Exec(p)

	// Must wait for exec program to finish before closing the container!
	result := <-resultChan
	if !errors.Is(result.Err, internal.ErrNotMounted) {
		t.Fatalf("expected error '%v', got: %v", internal.ErrNotMounted, result.Err)
	}
	if result.Status != -1 {
		t.Fatalf("expected exit code -1, got: %d", result.Status)
	}
	if result.ID != "" {
		t.Fatal("expected empty ID, got non-empty string")
	}

	close(stdinChan) // Allow container to close
	result = <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}
}

func TestExecMounts(t *testing.T) {
	stdinChan := make(chan string, 1)
	path := internal.NewTestDir(t).TouchFile("file.txt")
	c := internal.NewContainer(commonImage,
		internal.WithStdinChannel(stdinChan),
		internal.WithMounts(path))
	wait, err := program.FromSystem("cat")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := c.Run(wait); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, err := program.FromSystem("cat", program.WithFileArgs(path))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	resultChan := c.Exec(p)

	// Must wait for exec program to finish before closing the container!
	result := <-resultChan
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID == "" {
		t.Fatal("expected non-empty ID, got empty string")
	}

	close(stdinChan) // Allow container to close
	result = <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != c.ID {
		t.Fatalf("expected ID '%s', got: %s", c.ID, result.ID)
	}
}
