package stream

import (
	"manage2/internal"
)

type ChannelWriter struct {
	ch internal.DynamicChan[string]
}

func NewChannelWriter(target chan<- string) *ChannelWriter {
	if target == nil {
		panic("target channel cannot be nil")
	}
	cw := &ChannelWriter{
		ch: internal.NewDynamicChan[string](),
	}
	go func() {
		for value := range cw.ch.Pop {
			target <- value
		}
		close(target)
	}()
	return cw
}

func (cw *ChannelWriter) Write(p []byte) (n int, err error) {
	value := string(p)
	cw.ch.Push <- value
	return len(p), nil
}

func (cw *ChannelWriter) Close() error {
	close(cw.ch.Push)
	return nil
}
