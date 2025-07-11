package stream

type Option func(c *ContainerStream)

func WithStdinChannel(stdin <-chan string) Option {
	return func(c *ContainerStream) {
		c.Stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) Option {
	return func(c *ContainerStream) {
		c.Stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) Option {
	return func(c *ContainerStream) {
		c.Stderr = stderr
	}
}
