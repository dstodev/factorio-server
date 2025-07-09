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
	program, err := program.FromSystem("echo", program.WithStringArgs("Hello,", "world!"))

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

	select {
	case msg := <-stdoutChan:
		if msg != expectedOutput {
			t.Errorf("expected stdout message '%s', got: %s", expectedOutput, msg)
		}
	default:
		t.Error("expected a message on stdout")
	}
}
