package internal

type Queue[T any] struct {
	enqueue chan<- T
	dequeue <-chan T
}

func NewQueue[T any]() Queue[T] {
	enqueue := make(chan T)
	dequeue := make(chan T)

	q := Queue[T]{
		enqueue: enqueue,
		dequeue: dequeue,
	}

	go func() {
		var buffer []T
		var next T
		var outCh chan<- T

		defer func() {
			// Send all items before closing dequeue. This blocks until all
			// messages are either consumed, or stored in dequeue if it is
			// buffered.
			for _, item := range buffer {
				dequeue <- item
			}
			close(dequeue)
		}()

		for {
			if len(buffer) > 0 {
				next = buffer[0]
				outCh = dequeue
			} else {
				outCh = nil
			}

			select {
			case msg, ok := <-enqueue:
				if !ok {
					return
				}
				buffer = append(buffer, msg)

			case outCh <- next:
				// Do not remove from buffer until actually written. Prevents
				// loss of item if enqueue closes before writing it to dequeue.
				buffer = buffer[1:]
			}
		}
	}()
	return q
}

func (q Queue[T]) Close() {
	close(q.enqueue)
}

func (q Queue[T]) Enqueue(element T) {
	q.enqueue <- element
}

func (q Queue[T]) Dequeue() (value T, ok bool) {
	value, ok = <-q.dequeue
	return
}
