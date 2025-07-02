package internal_test

import (
	"testing"

	"manage2/internal"
)

func TestProgram(t *testing.T) {
	prog := internal.NewProgram("/usr/bin/someprogram", "alpine:latest")
	err := prog.Run()
	if err != nil {
		t.Fatalf("Failed to run program: %v", err)
	}
}
