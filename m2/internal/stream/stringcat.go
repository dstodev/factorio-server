package stream

import "strings"

type StringCat struct {
	Stdin chan<- string
	stdin <-chan string

	buffer strings.Builder
	done   chan struct{}
}

func NewStringCat() *StringCat {
	stdin := make(chan string)
	cat := &StringCat{
		Stdin: stdin,
		stdin: stdin,
		done:  make(chan struct{}),
	}
	cat.start()
	return cat
}

func (c *StringCat) start() {
	go func() {
		for line := range c.stdin {
			c.buffer.WriteString(line)
		}
		close(c.done)
	}()
}

func (c *StringCat) Print() string {
	<-c.done
	return c.buffer.String()
}
