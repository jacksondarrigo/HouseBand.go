package houseband

type Queue struct {
	items chan *songRequest
}

func (queue *Queue) enqueue(item *songRequest) {
	queue.items <- item
}

func (queue *Queue) dequeue() (item *songRequest) {
	item = <-queue.items
	return
}

func (queue *Queue) length() int {
	return len(queue.items)
}
