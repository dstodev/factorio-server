package program

type Program struct {
	Path string
	Args []string

	fileArgIndices IndexSet
}

func New(path string, args ...Argument) (*Program, error) {
	if err := checkFileExists(path); err != nil {
		return nil, err
	}
	p := &Program{
		Path:           path,
		Args:           nil,
		fileArgIndices: NewIndexSet(),
	}
	for _, arg := range args {
		if err := arg(p); err != nil {
			return nil, err
		}
	}
	return p, nil
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
