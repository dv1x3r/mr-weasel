package utils

import "context"

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

func (q *Queue) Lock(ctx context.Context) bool {
	select {
	case q.queueChan <- 0:
		select {
		case q.execChan <- 0:
			return true
		case <-ctx.Done():
			<-q.queueChan
			return false
		}
	default:
		return false
	}
}

func (q *Queue) Unlock() {
	<-q.queueChan
	<-q.execChan
}
