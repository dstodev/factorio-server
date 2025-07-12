package stream

import (
	"fmt"
	"io"
)

type ChannelWriter struct {
	target chan<- string
}

func NewChannelWriter(target chan<- string) io.Writer {
	if target == nil {
		return NewNilWriter()
	} else {
		return &ChannelWriter{
			target: target,
		}
	}
}

func (cw *ChannelWriter) Write(p []byte) (n int, err error) {
	if cw.target == nil {
		return 0, fmt.Errorf("target channel is nil")
	}
	value := string(p)
	cw.target <- value
	return len(p), nil
}

type NilWriter struct{}

func NewNilWriter() io.Writer {
	return &NilWriter{}
}

func (nw *NilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil // Discard the data
}
