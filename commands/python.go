package commands

import (
	"context"
	"sync"
	"time"

	"mr-weasel/utils"
)

type PythonCommand struct {
	queue *utils.Queue
}

func NewPythonCommand(queue *utils.Queue) *PythonCommand {
	return &PythonCommand{queue: queue}
}

func (PythonCommand) Prefix() string {
	return "/python"
}

func (PythonCommand) Description() string {
	return "test long running command"
}

func (c *PythonCommand) Execute(ctx context.Context, pl Payload) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if !c.queue.Lock() {
			pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
			return
		}
		defer c.queue.Unlock()
		time.Sleep(10 * time.Second)
		pl.ResultChan <- Result{Text: "Done, fuck you!"}
	}()

	pl.ResultChan <- Result{Text: "background job probably started"}
	wg.Wait()
}
