// Package queue implements a simple queue for uint64s - it is not thread safe (uncomment the mutex lines if you need thread safety)
package queue

// Pair holds the uint64 and an id
type Pair struct {
	Value uint64 // we use this for hashed k-mers
	ID    uint   // we use this for the position of the hashed k-mer in the sequence
}

// uint64Queue is a queue for items
type uint64Queue struct {
	//	sync.RWMutex
	items []Pair
}

// NewQueue is the constructor function for a uint64Queue
func NewQueue() *uint64Queue {
	return &uint64Queue{
		items: []Pair{},
	}
}

// Front is a method to return the Pair at the start of the queue, without removing it from the queue
func (uint64Queue *uint64Queue) Front() *Pair {
	//	uint64Queue.RLock()
	returnMinimizer := uint64Queue.items[0]
	//	uint64Queue.RUnlock()
	return &returnMinimizer
}

// Back is a method to return the Pair at the back of the queue, without removing it from the queue
func (uint64Queue *uint64Queue) Back() *Pair {
	//	uint64Queue.RLock()
	returnMinimizer := uint64Queue.items[len(uint64Queue.items)-1]
	//	uint64Queue.RUnlock()
	return &returnMinimizer
}

// PopFront is a method to remove and return the Pair at the start of the queue
func (uint64Queue *uint64Queue) PopFront() *Pair {
	//	uint64Queue.Lock()
	returnMinimizer := uint64Queue.items[0]
	uint64Queue.items = uint64Queue.items[1:len(uint64Queue.items)]
	//	uint64Queue.Unlock()
	return &returnMinimizer
}

// PopBack is a method to remove and return the Pair at the back of the queue
func (uint64Queue *uint64Queue) PopBack() *Pair {
	//	uint64Queue.Lock()
	returnMinimizer := uint64Queue.items[len(uint64Queue.items)-1]
	uint64Queue.items = uint64Queue.items[:len(uint64Queue.items)-1]
	//	uint64Queue.Unlock()
	return &returnMinimizer
}

// PushBack is a method to add a Pair to the back of the queue
func (uint64Queue *uint64Queue) PushBack(m Pair) {
	//	uint64Queue.Lock()
	uint64Queue.items = append(uint64Queue.items, m)
	//	uint64Queue.Unlock()
}

// Cardinality is a method to return the number of items currently in the queue
func (uint64Queue *uint64Queue) Cardinality() int {
	return len(uint64Queue.items)
}

// IsEmpty is a method to check if the queue currently has any items in it
func (uint64Queue *uint64Queue) IsEmpty() bool {
	return uint64Queue.Cardinality() == 0
}
