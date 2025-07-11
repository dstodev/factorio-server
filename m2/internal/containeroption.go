package internal

type ContainerOption func(c *Container)

func WithStdinChannel(stdin <-chan string) ContainerOption {
	return func(c *Container) {
		c.streams.Stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) ContainerOption {
	return func(c *Container) {
		c.streams.Stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) ContainerOption {
	return func(c *Container) {
		c.streams.Stderr = stderr
	}
}

func WithUser(user string) ContainerOption {
	return func(c *Container) {
		c.user = user
	}
}

func WithMounts(hostPaths ...string) ContainerOption {
	return func(c *Container) {
		for _, hostPath := range hostPaths {
			c.hostToGuestMap[hostPath] = toGuestPath(hostPath)
		}
	}
}
