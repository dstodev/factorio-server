package container

type Option func(ctr *Container)

func WithStdinChannel(stdin <-chan string) Option {
	return func(ctr *Container) {
		ctr.streams.Stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) Option {
	return func(ctr *Container) {
		ctr.streams.Stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) Option {
	return func(ctr *Container) {
		ctr.streams.Stderr = stderr
	}
}

func WithUser(user string) Option {
	return func(ctr *Container) {
		ctr.user = user
	}
}

func WithMounts(hostPaths ...string) Option {
	return func(ctr *Container) {
		for _, hostPath := range hostPaths {
			ctr.hostToGuestMap[hostPath] = toGuestPath(hostPath)
		}
	}
}
