package internal

type DynamicChan[T any] struct {
	Push chan<- T
	push <-chan T
	Pop  <-chan T
	pop  chan<- T
}

func NewDynamicChan[T any]() DynamicChan[T] {
	push := make(chan T)
	pop := make(chan T)

	d := DynamicChan[T]{
		Push: push,
		push: push,
		Pop:  pop,
		pop:  pop,
	}

	go func() {
		var buffer []T
		var next T
		var outCh chan<- T

		defer func() {
			for _, item := range buffer {
				d.pop <- item
			}
			close(d.pop)
		}()

		for {
			if len(buffer) > 0 {
				next = buffer[0]
				outCh = d.pop
			} else {
				outCh = nil
			}

			select {
			case msg, ok := <-d.push:
				if !ok {
					return
				}
				buffer = append(buffer, msg)

			case outCh <- next:
				buffer = buffer[1:]
			}
		}
	}()

	return d
}
