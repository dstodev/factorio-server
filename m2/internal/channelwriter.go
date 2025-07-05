package internal

import (
	"fmt"
	"io"
)

type ChannelWriter struct {
	targetChan chan<- string
}

func NewChannelWriter(targetChan chan<- string) io.Writer {
	if targetChan == nil {
		return NewNilWriter()
	} else {
		return &ChannelWriter{
			targetChan: targetChan,
		}
	}
}

func (cw *ChannelWriter) Write(p []byte) (n int, err error) {
	if cw.targetChan == nil {
		return 0, fmt.Errorf("target channel is nil")
	}
	value := string(p)
	cw.targetChan <- value
	return len(p), nil
}

type NilWriter struct{}

func NewNilWriter() io.Writer {
	return &NilWriter{}
}

func (nw *NilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil // Discard the data
}
