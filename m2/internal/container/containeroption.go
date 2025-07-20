package container

import "manage2/internal/container/stream"

type ContainerOption func(ctr *Container)

func WithStdinChannel(stdin <-chan string) ContainerOption {
	return func(ctr *Container) {
		stream.WithStdinChannel(stdin)(ctr.cs)
	}
}

func WithStdoutChannel(stdout chan<- string) ContainerOption {
	return func(ctr *Container) {
		stream.WithStdoutChannel(stdout)(ctr.cs)
	}
}

func WithStderrChannel(stderr chan<- string) ContainerOption {
	return func(ctr *Container) {
		stream.WithStderrChannel(stderr)(ctr.cs)
	}
}

func WithUser(user string) ContainerOption {
	return func(ctr *Container) {
		ctr.user = user
	}
}

func WithMounts(hostPaths ...string) ContainerOption {
	return func(ctr *Container) {
		for _, hostPath := range hostPaths {
			ctr.HostToGuestMap[hostPath] = toGuestPath(hostPath)
		}
	}
}
