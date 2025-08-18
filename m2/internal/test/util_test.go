package test_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"manage2/internal/test"
)

func TestWorkdir(t *testing.T) {
	workdir := test.Workdir(t)
	if workdir.Path == "" {
		t.Fatal("expected non-empty path for workdir")
	}
	if _, err := os.Stat(workdir.Path); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected workdir to exist: %s", workdir.Path)
	}
}

func TestTouchFile(t *testing.T) {
	workdir := test.Workdir(t)
	fileName := "testfile.txt"

	filePath := workdir.Path + "/" + fileName

	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file to not exist: %s", filePath)
	}

	touchPath := workdir.TouchFile(fileName)
	if touchPath != filePath {
		t.Fatalf("expected path '%s', received: %s", filePath, filePath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file '%s': %v", filePath, err)
	}
	if info.Mode() != 0o644 {
		t.Fatalf("expected file mode '%o', received: %o", 0o644, info.Mode())
	}
}

func TestWriteFile(t *testing.T) {
	workdir := test.Workdir(t)

	cases := []struct {
		name            string
		inputContent    string
		expectedContent string
		mode            os.FileMode
	}{
		{
			name:            "empty-file.txt",
			inputContent:    "",
			expectedContent: "",
			mode:            0o644,
		},
		{
			name:            "hello.txt",
			inputContent:    "Hello, world!",
			expectedContent: "Hello, world!\n",
			mode:            0o644,
		},
		{
			name:            "script.sh",
			inputContent:    "#!/bin/sh\necho 'Hello, world!'",
			expectedContent: "#!/bin/sh\necho 'Hello, world!'\n",
			mode:            0o755,
		},
		{
			name:            "newlines.txt",
			inputContent:    "\n\nLine 1\nLine 2\n\n",
			expectedContent: "Line 1\nLine 2\n",
			mode:            0o644,
		},
		{
			name:            "spaces.txt",
			inputContent:    "  \t  \r  \n  ",
			expectedContent: "",
			mode:            0o644,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			filePath := workdir.Path + "/" + c.name

			if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected file to not exist: %s", filePath)
			}

			writePath := workdir.WriteFile(c.name, c.inputContent, c.mode)
			if writePath != filePath {
				t.Fatalf("expected path '%s', received: %s", filePath, writePath)
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file '%s': %v", filePath, err)
			}

			if string(data) != c.expectedContent {
				expected := strings.ReplaceAll(c.expectedContent, "\n", `\n`)
				received := strings.ReplaceAll(string(data), "\n", `\n`)
				t.Fatalf("expected content '%s', received: %s", expected, received)
			}

			info, err := os.Stat(filePath)
			if err != nil {
				t.Fatalf("failed to stat file '%s': %v", filePath, err)
			}
			if c.mode != info.Mode() {
				t.Fatalf("expected file mode '%o', received: %o", c.mode, info.Mode())
			}
		})
	}
}
