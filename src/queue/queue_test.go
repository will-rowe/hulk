package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	queue := NewQueue()
	// add two items to the queue, giving them different positions so we can differentiate them
	queue.PushBack(Pair{Y: 1})
	if queue.IsEmpty() {
		t.Fatal("failed to add an item to queue")
	}
	queue.PushBack(Pair{Y: 2})
	if queue.Cardinality() != 2 {
		t.Fatal("there should be 2 items in the queue")
	}
	// check the ordering is correct
	if queue.Front().Y != 1 {
		t.Fatal("queue ordering is incorrect")
	}
	// check popping
	_ = queue.PopBack()
	if queue.Cardinality() != 1 {
		t.Fatal("failed to pop from the queue")
	}
	if queue.Front().Y != 1 {
		t.Fatal("popped from the front of the queue, should have been from the back")
	}
}
