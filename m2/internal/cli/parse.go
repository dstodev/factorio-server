package cli

func SplitDelimiters(args []string) {
	// TODO
}

func SplitNextDelimiter(args []string) (next []string, remaining []string) {
	next = Delimit(args)
	for i, arg := range args {
		if arg == "--" {
			next = args[:i+1]
			remaining = args[i+1:]

			break
		}
	}
	return
}

func Delimit(args []string) []string {
	sz := len(args)
	if sz > 0 && args[sz-1] != "--" {
		opts := make([]string, sz+1)
		copy(opts, args)
		opts[sz] = "--"
		return opts
	}
	return args
}

func Expand(arg string) []string {
	if len(arg) > 2 && arg[0] == '-' && arg[1] != '-' {
		opts := make([]string, len(arg)-1)
		for i := 1; i < len(arg); i++ {
			opt := arg[i : i+1]
			opts[i-1] = "-" + opt
		}
		return opts
	}
	return []string{arg}
}
