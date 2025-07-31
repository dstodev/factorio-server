package stream

import (
	"manage2/internal"
)

type ChannelWriter struct {
	ch internal.Queue[string]
}

func NewChannelWriter(target chan<- string) (writer *ChannelWriter, writeDone <-chan struct{}) {
	if target == nil {
		panic("target channel cannot be nil")
	}
	cw := &ChannelWriter{
		ch: internal.NewQueue[string](),
	}
	done := make(chan struct{})
	go func() {
		defer func() {
			close(target)
			close(done) // close done last
		}()
		for {
			value, ok := cw.ch.Dequeue()
			if !ok {
				break
			}
			target <- value
		}
	}()
	return cw, done
}

func (cw *ChannelWriter) Write(p []byte) (n int, err error) {
	value := string(p)
	cw.ch.Enqueue(value)
	return len(p), nil
}

func (cw *ChannelWriter) Close() error {
	cw.ch.Close()
	return nil
}
