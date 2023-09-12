package utils

type Queue struct {
	maxQueue    int
	maxParallel int
	queueChan   chan int
	execChan    chan int
}

func NewQueue(maxQueue int, maxParallel int) *Queue {
	return &Queue{
		maxQueue:    maxQueue,
		maxParallel: maxParallel,
		queueChan:   make(chan int, maxQueue),
		execChan:    make(chan int, maxParallel),
	}
}

func (q *Queue) Lock() bool {
	select {
	case q.queueChan <- 0:
		q.execChan <- 0
		return true
	default:
		return false
	}
}

func (q *Queue) Unlock() {
	<-q.queueChan
	<-q.execChan
}
