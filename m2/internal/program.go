package internal

import (
	"errors"
	"io/fs"
	"os"
)

type Program struct {
	Path string
	Args []string

	fileArgIndices indexSet
}

type ProgramArgument func(*Program) error

func NewProgram(path string, args ...ProgramArgument) (*Program, error) {
	if err := checkFileExists(path); err != nil {
		return nil, err
	}
	p := &Program{
		Path:           path,
		Args:           nil,
		fileArgIndices: newIndexSet(),
	}
	for _, arg := range args {
		if err := arg(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func checkFileExists(file string) error {
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// WithProgramArgs adds string arguments to pass to the program.
func WithStringArgs(args ...string) ProgramArgument {
	return func(p *Program) error {
		p.Args = append(p.Args, args...)
		return nil
	}
}

// WithFileArgs adds file paths as arguments to the program.
func WithFileArgs(files ...string) ProgramArgument {
	return func(p *Program) error {
		for _, file := range files {
			if err := checkFileExists(file); err != nil {
				return err
			}
		}
		for _, file := range files {
			argIndex := len(p.Args)
			p.Args = append(p.Args, file)
			p.fileArgIndices.Add(argIndex)
		}
		return nil
	}
}

func (p *Program) ArgIsFile(index int) bool {
	return p.fileArgIndices.Contains(index)
}

type ArgMutator func(p *Program, index int, arg string) string

func (p *Program) AsTokens(mutators ...ArgMutator) []string {
	numTokens := len(p.Args) + 2 // +1 for the program path, +1 for the "--" terminator
	tokens := make([]string, numTokens)
	tokens[0] = p.Path
	tokens[numTokens-1] = "--"
	for i, arg := range p.Args {
		postProcessed := arg
		for _, mutate := range mutators {
			postProcessed = mutate(p, i, postProcessed)
		}
		tokens[i+1] = postProcessed // +1 to offset for the program path
	}
	return tokens
}
