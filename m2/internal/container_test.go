package internal_test

import (
	"errors"
	"testing"

	"manage2/internal"
	"manage2/internal/program"
)

func TestNewContainer(t *testing.T) {
	image := "some-image:latest"
	c := internal.NewContainer(image)

	if c.Image != image {
		t.Errorf("expected image '%s', got: %s", image, c.Image)
	}

	id, err := c.Run()
	if !errors.Is(err, internal.ErrNoProgram) {
		t.Errorf("expected error '%v', got: %v", internal.ErrNoProgram, err)
	}

	if id != "" {
		t.Errorf("expected empty ID, got: %s", id)
	}
}

func TestContainerWithProgram(t *testing.T) {
	image := "alpine:latest"
	program, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "exit 5"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	c := internal.NewContainer(image, internal.WithProgram(program))

	id, err := c.Run()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 5 {
		t.Errorf("expected exit code 5, got: %d", result.Status)
	}
	if result.ID != id {
		t.Errorf("expected ID '%s', got: %s", id, result.ID)
	}
}

func TestContainerWithStdout(t *testing.T) {
	image := "alpine:latest"
	program, err := program.FromSystem("printf", program.WithStringArgs("|%s|\n", "Hello,", "world!"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdoutChan := make(chan string, 1)

	c := internal.NewContainer(image,
		internal.WithProgram(program),
		internal.WithStdoutChannel(stdoutChan),
	)

	id, err := c.Run()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}

	expectedOutput := "|Hello,|\n|world!|\n|--|\n"

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Errorf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != id {
		t.Errorf("expected ID '%s', got: %s", id, result.ID)
	}

	select {
	case msg := <-stdoutChan:
		if msg != expectedOutput {
			t.Errorf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Error("expected a message on stdout")
	}
}

func TestContainerWithStdoutAndStderr(t *testing.T) {
	image := "alpine:latest"
	program, err := program.FromSystem("/bin/sh",
		program.WithStringArgs("-c", "printf '|%s|\n' Hello, world! | tee /dev/stderr"))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdoutChan := make(chan string, 1)
	stderrChan := make(chan string, 1)

	c := internal.NewContainer(image,
		internal.WithProgram(program),
		internal.WithStdoutChannel(stdoutChan),
		internal.WithStderrChannel(stderrChan),
	)

	id, err := c.Run()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}

	expectedOutput := "|Hello,|\n|world!|\n"

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Errorf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != id {
		t.Errorf("expected ID '%s', got: %s", id, result.ID)
	}

	select {
	case msg := <-stdoutChan:
		if msg != expectedOutput {
			t.Errorf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Error("expected a message on stdout")
	}

	select {
	case msg := <-stderrChan:
		if msg != expectedOutput {
			t.Errorf("expected stderr message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Error("expected a message on stderr")
	}
}

func TestContainerWithStdin(t *testing.T) {
	image := "alpine:latest"
	program, err := program.FromSystem("/bin/sh", program.WithStringArgs("-c", "read input && echo \"|$input|\""))

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdinChan := make(chan string, 1)
	stdoutChan := make(chan string, 1)

	c := internal.NewContainer(image,
		internal.WithProgram(program),
		internal.WithStdinChannel(stdinChan),
		internal.WithStdoutChannel(stdoutChan),
	)

	msg := "Hello, world!\n"
	expectedMsg := "|Hello, world!|\n"

	stdinChan <- msg
	close(stdinChan)

	id, err := c.Run()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}

	result := <-c.Done
	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Status != 0 {
		t.Errorf("expected exit code 0, got: %d", result.Status)
	}
	if result.ID != id {
		t.Errorf("expected ID '%s', got: %s", id, result.ID)
	}

	select {
	case msg := <-stdoutChan:
		if msg != expectedMsg {
			t.Errorf("expected stdout message '%s', got: %s", expectedMsg, msg)
		}
	default:
		t.Error("expected a message on stdout")
	}
}
