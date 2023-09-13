package commands

import (
	"context"
	"os/exec"
	"sync"

	"mr-weasel/utils"
)

type SeparateSongCommand struct {
	queue *utils.Queue
}

func NewSeparateSongCommand(queue *utils.Queue) *SeparateSongCommand {
	return &SeparateSongCommand{queue: queue}
}

func (SeparateSongCommand) Prefix() string {
	return "/separatesong"
}

func (SeparateSongCommand) Description() string {
	return "separate voice from music out of a song"
}

func (c *SeparateSongCommand) Execute(ctx context.Context, pl Payload) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if c.queue.Lock() {
			defer c.queue.Unlock()
			c.python(pl)
		} else {
			pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
		}
	}()

	pl.ResultChan <- Result{Text: "background job probably started"}
	wg.Wait()
}

func (c *SeparateSongCommand) python(pl Payload) {
	path, err := exec.LookPath("sleep")
	if err != nil {
		pl.ResultChan <- Result{Text: "Command failed successfully!", Error: err}
	}
	cmd := exec.Command(path, "5")
	err = cmd.Run()
	if err != nil {
		pl.ResultChan <- Result{Text: "Command failed successfully!", Error: err}
	} else {
		pl.ResultChan <- Result{Text: "Done!"}
	}
}
