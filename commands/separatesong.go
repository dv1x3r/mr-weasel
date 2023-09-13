package commands

import (
	"context"
	"os"
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
	c.startProcessing(ctx, pl, "smash.mp3")
}

func (c *SeparateSongCommand) startProcessing(ctx context.Context, pl Payload, fileName string) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if c.queue.Lock() {
			defer c.queue.Unlock()
			if err := c.runPython(ctx, fileName); err != nil {
				pl.ResultChan <- Result{Text: "Failed!", Error: err}
			} else {
				pl.ResultChan <- Result{Text: "Done!"}
			}
		} else {
			pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
		}
	}()

	pl.ResultChan <- Result{Text: "background job probably started"}
	wg.Wait()

}

func (c *SeparateSongCommand) runPython(ctx context.Context, fileName string) error {
	cmd := exec.CommandContext(
		ctx,
		"/home/dx/source/audio-separator/.venv/bin/python",
		"/home/dx/source/audio-separator/cli.py",
		"/home/dx/source/audio-separator/tracks/"+fileName,
		"--model_file_dir=/home/dx/source/audio-separator/models/",
		"--output_dir=/home/dx/source/audio-separator/tracks/",
		"--use_cuda",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
