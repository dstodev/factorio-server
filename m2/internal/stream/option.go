package stream

type Option func(c *ContainerStream)

func WithStdinChannel(stdin <-chan string) Option {
	return func(cs *ContainerStream) {
		cs.Stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) Option {
	return func(cs *ContainerStream) {
		cs.Stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) Option {
	return func(cs *ContainerStream) {
		cs.Stderr = stderr
	}
}
