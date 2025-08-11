package internal_test

import (
	"manage2/internal"
	"sync"
	"testing"
)

func TestQueue(t *testing.T) {
	q := internal.NewQueue[int]()

	numValues := 100

	var wgDone, wgStart sync.WaitGroup
	wgDone.Add(numValues)
	wgStart.Add(numValues)

	go func() {
		wgDone.Wait()
		q.Close()
	}()
	barrier := make(chan struct{})
	for i := range numValues {
		go func() {
			defer wgDone.Done()
			wgStart.Done()
			<-barrier
			q.Enqueue(i)
		}()
	}
	wgStart.Wait()
	close(barrier)

	count := 0
	for range numValues {
		value, ok := q.Dequeue()
		if !ok {
			t.Errorf("Queue closed before all values were dequeued")
			return
		}
		count += value
	}

	// arithmetic series sum 0,1,...,numValues-1
	// sum = (n / 2) * (first + last)
	// n=numValues, first=0, last=numValues-1

	sum := (numValues * (numValues - 1)) / 2
	if count != sum {
		t.Errorf("Expected count %d, received %d", sum, count)
	}
	if value, ok := q.Dequeue(); ok {
		t.Error("Expected closed channel, received:", value)
	}
}
