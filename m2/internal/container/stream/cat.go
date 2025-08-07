package stream

import "strings"

type Cat struct {
	Stdin chan<- string
	stdin <-chan string

	buffer strings.Builder
	done   chan struct{}
}

func NewCat() *Cat {
	stdin := make(chan string)
	cat := &Cat{
		Stdin: stdin,
		stdin: stdin,
		done:  make(chan struct{}),
	}
	cat.start()
	return cat
}

func (c *Cat) start() {
	go func() {
		for line := range c.stdin {
			c.buffer.WriteString(line)
		}
		close(c.done)
	}()
}

func (c *Cat) String() string {
	<-c.done
	return c.buffer.String()
}
