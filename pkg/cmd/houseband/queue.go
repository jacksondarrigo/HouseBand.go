package houseband

type Queue struct {
	items []*songRequest
}

func (queue *Queue) enqueue(item *songRequest) {
	queue.items = append(queue.items, item)
}

func (queue *Queue) dequeue() (item *songRequest) {
	item = queue.items[0]
	if len(queue.items) > 1 {
		queue.items = queue.items[1:]
	}
	return
}

func (queue *Queue) length() int {
	return len(queue.items)
}
