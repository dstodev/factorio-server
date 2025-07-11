package stream

import (
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerStream struct {
	Stdin  <-chan string
	Stdout chan<- string
	Stderr chan<- string

	ctrSock *types.HijackedResponse // Container socket
	wg      *sync.WaitGroup
}

func NewContainerStream(opts ...Option) *ContainerStream {
	cs := &ContainerStream{
		wg: &sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(cs)
	}
	return cs
}

func (cs *ContainerStream) StartStreaming(ctrSock *types.HijackedResponse) {
	cs.ctrSock = ctrSock

	attachStdin := cs.Stdin != nil
	attachStdout := cs.Stdout != nil
	attachStderr := cs.Stderr != nil

	if attachStdout || attachStderr {
		cs.startWritingOutputToChannels(attachStdout, attachStderr)
	}

	if attachStdin {
		cs.startReadingInputFromChannel()
	} else {
		cs.ctrSock.CloseWrite()
	}
}

// startWritingOutputToChannels reads from the container's stdout and stderr,
// writing them to the respective channels. Channels are closed when the
// container's respective output streams close, if requested. If a channel is
// nil, output is still accepted from the container, then discarded.
func (cs *ContainerStream) startWritingOutputToChannels(
	closeStdout bool,
	closeStderr bool,
) {
	cs.wg.Add(1)

	go func() {
		defer cs.wg.Done()

		// Close output channels when container outputs close
		if closeStdout {
			defer close(cs.Stdout)
		}
		if closeStderr {
			defer close(cs.Stderr)
		}

		stdcopy.StdCopy(
			NewChannelWriter(cs.Stdout),
			NewChannelWriter(cs.Stderr),
			cs.ctrSock.Reader)
	}()
}

// startReadingInputFromChannel reads from the stdin channel, writing strings
// to the container's stdin.
func (cs *ContainerStream) startReadingInputFromChannel() {
	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		defer cs.ctrSock.CloseWrite() // Close write after stdin channel closes

		for input := range cs.Stdin {
			if _, err := cs.ctrSock.Conn.Write([]byte(input)); err != nil {
				return
			}
		}
	}()
}

// WaitUntilClosed waits for all streams to close, then cleans up.
func (cs *ContainerStream) WaitUntilClosed() {
	defer func() {
		if cs.ctrSock != nil {
			cs.ctrSock.Close()
		}
	}()
	cs.wg.Wait()
}
