package program

type Program struct {
	Args []string

	fileArgIndices Set[int]
}

func FromSystem(path string, args ...Argument) (*Program, error) {
	args = append([]Argument{WithStringArgs(path)}, args...) // Prepend the program path
	return fromArgs(args)
}

func fromArgs(args []Argument) (*Program, error) {
	p := &Program{
		Args:           nil,
		fileArgIndices: NewSet[int](),
	}
	for _, arg := range args {
		if err := arg(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func FromFile(path string, args ...Argument) (*Program, error) {
	args = append([]Argument{WithFileArgs(path)}, args...) // Prepend the program path
	return fromArgs(args)
}

func (p *Program) ArgIsFile(index int) bool {
	return p.fileArgIndices.Contains(index)
}

type ArgMutator func(p *Program, index int, arg string) string

func (p *Program) AsTokens(mutators ...ArgMutator) []string {
	numTokens := len(p.Args) + 1 // +1 for the "--" terminator
	tokens := make([]string, numTokens)
	tokens[numTokens-1] = "--"
	for i, arg := range p.Args {
		postProcessed := arg
		for _, mutate := range mutators {
			postProcessed = mutate(p, i, postProcessed)
		}
		tokens[i] = postProcessed
	}
	return tokens
}
