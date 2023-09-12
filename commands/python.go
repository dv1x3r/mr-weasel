package commands

import (
	"context"
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

func (c *PythonCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	resultChan := make(chan Result)
	go func() {
		if !c.queue.Lock() {
			resultChan <- Result{Text: "There are too many queued jobs, please wait."}
			return
		}
		defer c.queue.Unlock()
		time.Sleep(10 * time.Second)
		resultChan <- Result{Text: "Done, fuck you!"}
	}()

	return Result{Text: "background job probably started", ResultChan: resultChan}, nil
}
