package stream

import (
	"io"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerStream struct {
	Stdin  <-chan string
	Stdout chan<- string
	Stderr chan<- string

	stdoutWriter io.WriteCloser
	stderrWriter io.WriteCloser

	ctrSock *types.HijackedResponse // Container socket
	wg      *sync.WaitGroup
}

func NewContainerStream(opts ...StreamOption) *ContainerStream {
	cs := &ContainerStream{
		wg: &sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(cs)
	}
	return cs
}

// StartStreaming starts the container stream, facilitating communication over
// the container's stdin, stdout, and stderr. Once output streams are closed
// with CloseOutputChannels(), a ContainerStream is no longer usable.
func (cs *ContainerStream) StartStreaming(ctrSock *types.HijackedResponse) {
	cs.ctrSock = ctrSock

	attachStdin := cs.Stdin != nil
	attachStdout := cs.Stdout != nil
	attachStderr := cs.Stderr != nil

	if attachStdout || attachStderr {
		cs.startWritingOutputToChannels()
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
func (cs *ContainerStream) startWritingOutputToChannels() {
	stdout := io.Discard
	stderr := io.Discard

	if cs.Stdout != nil {
		cs.stdoutWriter = NewChannelWriter(cs.Stdout)
		stdout = cs.stdoutWriter
	}
	if cs.Stderr != nil {
		cs.stderrWriter = NewChannelWriter(cs.Stderr)
		stderr = cs.stderrWriter
	}

	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		stdcopy.StdCopy(stdout, stderr, cs.ctrSock.Reader)
	}()
}

// startReadingInputFromChannel reads from the stdin channel, writing strings
// to the container's stdin.
func (cs *ContainerStream) startReadingInputFromChannel() {
	cs.wg.Add(1)
	go func() {
		defer func() {
			cs.ctrSock.CloseWrite()
			cs.wg.Done()
		}()

		for input := range cs.Stdin {
			if _, err := cs.ctrSock.Conn.Write([]byte(input)); err != nil {
				return
			}
		}
	}()
}

func (cs *ContainerStream) Close() {
	cs.wg.Wait()
	if cs.ctrSock != nil {
		cs.ctrSock.Close()
	}
	cs.closeOutputChannels()
}

func (cs *ContainerStream) closeOutputChannels() {
	if cs.stdoutWriter != nil {
		cs.stdoutWriter.Close()
	}
	if cs.stderrWriter != nil {
		cs.stderrWriter.Close()
	}
}
