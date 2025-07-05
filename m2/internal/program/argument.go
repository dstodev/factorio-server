package program

// An Argument is a function which adds an argument to the program.
// This enables complex argument types, like file paths, to record
// extra data to the program for later use.
type Argument func(program *Program) error

// WithProgramArgs adds string arguments to the program.
func WithStringArgs(args ...string) Argument {
	return func(p *Program) error {
		p.Args = append(p.Args, args...)
		return nil
	}
}

// WithFileArgs adds file paths as arguments to the program.
func WithFileArgs(files ...string) Argument {
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
