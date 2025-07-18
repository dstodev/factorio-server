package container

import "manage2/internal/stream"

type Option func(ctr *Container)

func WithStdinChannel(stdin <-chan string) Option {
	return func(ctr *Container) {
		stream.WithStdinChannel(stdin)(ctr.streams)
	}
}

func WithStdoutChannel(stdout chan<- string) Option {
	return func(ctr *Container) {
		stream.WithStdoutChannel(stdout)(ctr.streams)
	}
}

func WithStderrChannel(stderr chan<- string) Option {
	return func(ctr *Container) {
		stream.WithStderrChannel(stderr)(ctr.streams)
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
			ctr.HostToGuestMap[hostPath] = toGuestPath(hostPath)
		}
	}
}
