package queue

import "context"

type Queue struct {
	queueChan chan struct{}
	execChan  chan struct{}
}

func NewQueue(queueBuffer, execBuffer int) *Queue {
	return &Queue{
		queueChan: make(chan struct{}, queueBuffer),
		execChan:  make(chan struct{}, execBuffer),
	}
}

func (q *Queue) Lock(ctx context.Context) bool {
	select {
	case q.queueChan <- struct{}{}:
		select {
		case q.execChan <- struct{}{}:
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
