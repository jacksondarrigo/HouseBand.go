package houseband

type queue struct {
	queue []request
}

func (q *queue) enqueue(request request) {
	q.queue = append(q.queue, request)
}

func (q *queue) dequeue() request {
	request := q.queue[0]
	q.queue = q.queue[1:]
	return request
}

func (q *queue) length() int {
	return len(q.queue)
}

func (q *queue) isEmpty() bool {
	return len(q.queue) < 1
}
