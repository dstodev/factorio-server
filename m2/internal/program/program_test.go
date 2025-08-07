package program_test

import (
	"errors"
	"io/fs"
	"testing"

	"manage2/internal/program"
	"manage2/internal/test"
)

func TestNewProgram(t *testing.T) {
	path := test.Workdir(t).TouchFile("test.sh")
	pgm, err := program.FromFile(path)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}
	if len(pgm.Args) != 1 {
		t.Fatalf("expected 1 arg, received %d", len(pgm.Args))
	}
}

func TestNewProgramBadPath(t *testing.T) {
	path := test.Workdir(t).Path + "/test.sh"
	pgm, err := program.FromFile(path)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected error to be 'fs.ErrNotExist', received %v", err)
	}
	if pgm != nil {
		t.Fatalf("expected nil program, received %v", pgm)
	}
}

func TestFromSystemWithArgs(t *testing.T) {
	pgm, err := program.FromSystem("echo", program.WithStringArgs("Hello,", "World!"))
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}

	expectedArgs := []string{"echo", "Hello,", "World!"}
	if len(pgm.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, received %d", len(expectedArgs), len(pgm.Args))
	}
	for i, arg := range expectedArgs {
		if pgm.Args[i] != arg {
			t.Fatalf("expected arg %d to be '%s', received %s", i, arg, pgm.Args[i])
		}
		if pgm.ArgIsFile(i) {
			t.Fatalf("expected arg %d to not be a file", i)
		}
	}
}

func TestFromFileWithArgs(t *testing.T) {
	path := test.Workdir(t).TouchFile("test.sh")

	pgm, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
	)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}

	expectedArgs := []string{path, "arg1", "arg2"}
	if len(pgm.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, received %d", len(expectedArgs), len(pgm.Args))
	}
	for i, arg := range expectedArgs {
		if pgm.Args[i] != arg {
			t.Fatalf("expected arg %d to be '%s', received %s", i, arg, pgm.Args[i])
		}
	}
}

func TestNewProgramWithFileArgs(t *testing.T) {
	dir := test.Workdir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	pgm, err := program.FromFile(path,
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}

	expectedArgs := []string{path, file1, file2}
	if len(pgm.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, received %d", len(expectedArgs), len(pgm.Args))
	}
	for i, arg := range expectedArgs {
		if pgm.Args[i] != arg {
			t.Fatalf("expected arg %d to be '%s', received %s", i, arg, pgm.Args[i])
		}
	}
}

func TestNewProgramWithBadFileArgs(t *testing.T) {
	path := test.Workdir(t).TouchFile("test.sh")

	_, err := program.FromFile(path,
		program.WithFileArgs("file.txt"),
	)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected error to be 'fs.ErrNotExist', received %v", err)
	}
}

func TestArgIsFile(t *testing.T) {
	dir := test.Workdir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	pgm, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}
	if !pgm.ArgIsFile(0) {
		t.Fatalf("expected arg 0 to be a file")
	}
	if pgm.ArgIsFile(1) || pgm.ArgIsFile(2) {
		t.Fatalf("expected args 0 and 1 to not be files")
	}
	if !pgm.ArgIsFile(3) || !pgm.ArgIsFile(4) {
		t.Fatalf("expected args 2 and 3 to be files")
	}
}

func TestAsTokens(t *testing.T) {
	dir := test.Workdir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	pgm, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}

	tokens := pgm.AsTokens()
	expectedTokens := []string{path, "arg1", "arg2", file1, file2, "--"}
	if len(tokens) != len(expectedTokens) {
		t.Fatalf("expected %d tokens, received %d", len(expectedTokens), len(tokens))
	}
	for i, expectedToken := range expectedTokens {
		if tokens[i] != expectedToken {
			t.Fatalf("expected token %d to be '%s', received %s", i, expectedToken, tokens[i])
		}
	}
}

func TestAsTokenMutators(t *testing.T) {
	dir := test.Workdir(t)
	path := dir.TouchFile("test.sh")
	file1 := dir.TouchFile("file1.txt")
	file2 := dir.TouchFile("file2.txt")

	pgm, err := program.FromFile(path,
		program.WithStringArgs("arg1", "arg2"),
		program.WithFileArgs(file1, file2),
	)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}

	count := 0

	mutator1 := func(pgm *program.Program, index int, arg string) string {
		count += 1
		if pgm.ArgIsFile(index) {
			return "file:" + arg
		}
		return "string:" + arg
	}
	mutator2 := func(pgm *program.Program, index int, arg string) string {
		if index == 0 {
			return "base:" + arg
		}
		return arg
	}

	tokens := pgm.AsTokens(mutator1, mutator2)

	expectedCount := len(pgm.Args) // mutators not called for "--"

	if count != expectedCount {
		t.Fatalf("expected %d calls to mutators, received %d", expectedCount, count)
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
		t.Fatalf("expected %d tokens, received %d", len(expectedTokens), len(tokens))
	}
	for i, expectedToken := range expectedTokens {
		if tokens[i] != expectedToken {
			t.Fatalf("expected token %d to be '%s', received %s", i, expectedToken, tokens[i])
		}
	}
}
