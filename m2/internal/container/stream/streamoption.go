package stream

type StreamOption func(c *ContainerStream)

func WithStdinChannel(stdin <-chan string) StreamOption {
	return func(cs *ContainerStream) {
		cs.Stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) StreamOption {
	return func(cs *ContainerStream) {
		cs.Stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) StreamOption {
	return func(cs *ContainerStream) {
		cs.Stderr = stderr
	}
}
