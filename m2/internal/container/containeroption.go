package container

type Option func(c *Container)

func WithStdinChannel(stdin <-chan string) Option {
	return func(c *Container) {
		c.streams.Stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) Option {
	return func(c *Container) {
		c.streams.Stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) Option {
	return func(c *Container) {
		c.streams.Stderr = stderr
	}
}

func WithUser(user string) Option {
	return func(c *Container) {
		c.user = user
	}
}

func WithMounts(hostPaths ...string) Option {
	return func(c *Container) {
		for _, hostPath := range hostPaths {
			c.hostToGuestMap[hostPath] = toGuestPath(hostPath)
		}
	}
}
