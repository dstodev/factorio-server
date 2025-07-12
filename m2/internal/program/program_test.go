package program_test

import (
	"errors"
	"io/fs"
	"testing"

	"manage2/internal"
	"manage2/internal/program"
)

func TestNewProgram(t *testing.T) {
	path := internal.NewTestDir(t).TouchFile("test.sh")
	p, err := program.FromFile(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(p.Args) != 1 {
		t.Fatalf("expected 1 arg, got: %d", len(p.Args))
	}
}

func TestNewProgramBadPath(t *testing.T) {
	path := internal.NewTestDir(t).Path + "/test.sh"
	p, err := program.FromFile(path)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected error to be 'fs.ErrNotExist', got: %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil program, got: %v", p)
	}
}

func TestFromSystemWithArgs(t *testing.T) {
	p, err := program.FromSystem("echo", program.WithStringArgs("Hello,", "World!"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedArgs := []string{"echo", "Hello,", "World!"}
	if len(p.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, got: %d", len(expectedArgs), len(p.Args))
	}
	for i, arg := range expectedArgs {
		if p.Args[i] != arg {
			t.Fatalf("expected arg %d to be '%s', got: %s", i, arg, p.Args[i])
		}
		if p.ArgIsFile(i) {
			t.Fatalf("expected arg %d to not be a file", i)
		}
	}
}

func TestFromFileWithArgs(t *testing.T) {
	path := internal.NewTestDir(t).TouchFile("test.sh")

	p, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedArgs := []string{path, "arg1", "arg2"}
	if len(p.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, got: %d", len(expectedArgs), len(p.Args))
	}
	for i, arg := range expectedArgs {
		if p.Args[i] != arg {
			t.Fatalf("expected arg %d to be '%s', got: %s", i, arg, p.Args[i])
		}
	}
}

func TestNewProgramWithFileArgs(t *testing.T) {
	dir := internal.NewTestDir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	p, err := program.FromFile(path,
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedArgs := []string{path, file1, file2}
	if len(p.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, got: %d", len(expectedArgs), len(p.Args))
	}
	for i, arg := range expectedArgs {
		if p.Args[i] != arg {
			t.Fatalf("expected arg %d to be '%s', got: %s", i, arg, p.Args[i])
		}
	}
}

func TestNewProgramWithBadFileArgs(t *testing.T) {
	path := internal.NewTestDir(t).TouchFile("test.sh")

	_, err := program.FromFile(path,
		program.WithFileArgs("file.txt"),
	)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected error to be 'fs.ErrNotExist', got: %v", err)
	}
}

func TestArgIsFile(t *testing.T) {
	dir := internal.NewTestDir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	p, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !p.ArgIsFile(0) {
		t.Fatalf("expected arg 0 to be a file")
	}
	if p.ArgIsFile(1) || p.ArgIsFile(2) {
		t.Fatalf("expected args 0 and 1 to not be files")
	}
	if !p.ArgIsFile(3) || !p.ArgIsFile(4) {
		t.Fatalf("expected args 2 and 3 to be files")
	}
}

func TestAsTokens(t *testing.T) {
	dir := internal.NewTestDir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	p, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	tokens := p.AsTokens()
	expectedTokens := []string{path, "arg1", "arg2", file1, file2, "--"}
	if len(tokens) != len(expectedTokens) {
		t.Fatalf("expected %d tokens, got: %d", len(expectedTokens), len(tokens))
	}
	for i, expectedToken := range expectedTokens {
		if tokens[i] != expectedToken {
			t.Fatalf("expected token %d to be '%s', got: %s", i, expectedToken, tokens[i])
		}
	}
}

func TestAsTokenMutators(t *testing.T) {
	dir := internal.NewTestDir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	p, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	count := 0

	mutator1 := func(p *program.Program, index int, arg string) string {
		count += 1
		if p.ArgIsFile(index) {
			return "file:" + arg
		}
		return "string:" + arg
	}
	mutator2 := func(p *program.Program, index int, arg string) string {
		if index == 0 {
			return "base:" + arg
		}
		return arg
	}

	tokens := p.AsTokens(mutator1, mutator2)

	expectedCount := len(p.Args) // mutators not called for "--"

	if count != expectedCount {
		t.Fatalf("expected %d calls to mutators, got: %d", expectedCount, count)
	}

	expectedTokens := []string{
		"base:file:" + path,
		"string:arg1",
		"string:arg2",
		"file:" + file1,
		"file:" + file2,
		"--",
	}

	if len(tokens) != len(expectedTokens) {
		t.Fatalf("expected %d tokens, got: %d", len(expectedTokens), len(tokens))
	}
	for i, expectedToken := range expectedTokens {
		if tokens[i] != expectedToken {
			t.Fatalf("expected token %d to be '%s', got: %s", i, expectedToken, tokens[i])
		}
	}
}
