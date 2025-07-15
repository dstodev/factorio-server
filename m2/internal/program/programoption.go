package program

import (
	"os"
)

// A Option is a function which adds an argument to the program.
// This enables complex argument types, like file paths, to record
// extra data to the program for later use.
type Option func(pgm *Program) error

// WithStringArgs adds string arguments to the program.
func WithStringArgs(args ...string) Option {
	return func(pgm *Program) error {
		pgm.Args = append(pgm.Args, args...)
		return nil
	}
}

// WithFileArgs adds file paths as arguments to the program.
func WithFileArgs(files ...string) Option {
	return func(pgm *Program) error {
		for _, file := range files {
			if err := checkFileExists(file); err != nil {
				return err
			}
		}
		for _, file := range files {
			argIndex := len(pgm.Args)
			pgm.Args = append(pgm.Args, file)
			pgm.fileArgIndices.Add(argIndex)
		}
		return nil
	}
}

func checkFileExists(file string) error {
	_, err := os.Stat(file)
	return err
}
