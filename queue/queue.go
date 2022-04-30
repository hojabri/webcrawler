package queue

import (
	"sync"
)

const chunkSize = 128

// chunks are used to make a queue auto resizeable.
type chunk struct {
	items       [chunkSize]any // list of queued items
	first, last int            // positions for the first and last item in this chunk
	next        *chunk         // pointer to the next chunk (if any)
}

type Queue struct {
	head, tail *chunk     // chunk head and tail
	count      int        // total amount of items in the queue
	lock       sync.Mutex // synchronisation lock
}

// NewQueue creates a new and empty *fifo.Queue
func NewQueue() (q *Queue) {
	initChunk := new(chunk)
	q = &Queue{
		head: initChunk,
		tail: initChunk,
	}
	return q
}

// Len returns the number of items in the queue
func (q *Queue) Len() (length int) {
	// locking to make Queue thread-safe
	q.lock.Lock()
	defer q.lock.Unlock()

	// copy q.count and return length
	length = q.count
	return length
}

// Add adds an item to the end of the queue
func (q *Queue) Add(item any) {
	// locking to make Queue thread-safe
	q.lock.Lock()
	defer q.lock.Unlock()

	// check if item is valid
	if item == nil {
		panic("can not add nil item to queue")
	}

	// if the tail chunk is full, create a new one and add it to the queue.
	if q.tail.last >= chunkSize {
		q.tail.next = new(chunk)
		q.tail = q.tail.next
	}

	// add item to the tail chunk at the last position
	q.tail.items[q.tail.last] = item
	q.tail.last++
	q.count++
}

// Pop Remove the item at the head of the queue and return it.
// Returns nil when there are no items left in queue.
func (q *Queue) Pop() (item any) {
	// locking to make Queue thread-safe
	q.lock.Lock()
	defer q.lock.Unlock()

	// Return nil if there are no items to return
	if q.count == 0 {
		return nil
	}

	// Get item from queue
	item = q.head.items[q.head.first]

	// increment first position and decrement queue item count
	q.head.first++
	q.count--

	if q.head.first >= q.head.last {
		// we're at the end of this chunk, and we should do some maintenance
		// if there are no follow-up chunks then reset the current one, so it can be used again.
		if q.count == 0 {
			q.head.first = 0
			q.head.last = 0
			q.head.next = nil
		} else {
			// set queue's head chunk to the next chunk
			// old head will fall out of scope and be GC-ed
			q.head = q.head.next
		}
	}

	// return the retrieved item
	return item
}
