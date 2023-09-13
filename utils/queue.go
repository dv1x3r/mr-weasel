package utils

type Queue struct {
	queueChan chan int
	execChan  chan int
}

func NewQueue(queueBuffer, execBuffer int) *Queue {
	return &Queue{
		queueChan: make(chan int, queueBuffer),
		execChan:  make(chan int, execBuffer),
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
