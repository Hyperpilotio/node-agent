package queue

import log "github.com/sirupsen/logrus"

// Queue is a FIFO (First in first out) data structure implementation.
// It is based on a deque container and focuses its API on core
// functionalities: Enqueue, Dequeue, Head, Size, Empty. Every operations time complexity
// is O(1).
//
// As it is implemented using a Deque container, every operations
// over an instiated Queue are synchronized and safe for concurrent
// usage.
type Queue struct {
	*Deque
}

func NewCappedQueue(capacity int) *Queue {
	return &Queue{
		Deque: NewCappedDeque(capacity),
	}
}

func NewQueue() *Queue {
	return &Queue{
		Deque: NewDeque(),
	}
}

// Enqueue adds an item at the back of the queue,
// but remove the first element when queue is full
func (q *Queue) Enqueue(item interface{}) {
	for !q.Prepend(item) {
		log.Warnf("Enqueue fail due to full queue, remove oldest metric")
		q.Pop()
	}
}

// Dequeue removes and returns the front queue item
func (q *Queue) Dequeue() interface{} {
	return q.Pop()
}

// Head returns the front queue item
func (q *Queue) Head() interface{} {
	return q.Last()
}
