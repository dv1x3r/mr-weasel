package commands

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"mr-weasel/utils"
)

type SeparateSongCommand struct {
	blob  *utils.Blob
	queue *utils.Queue
}

func NewSeparateSongCommand(blob *utils.Blob, queue *utils.Queue) *SeparateSongCommand {
	return &SeparateSongCommand{
		blob:  blob,
		queue: queue,
	}
}

func (SeparateSongCommand) Prefix() string {
	return "/separatesong"
}

func (SeparateSongCommand) Description() string {
	return "separate voice from music"
}

func (c *SeparateSongCommand) Execute(ctx context.Context, pl Payload) {
	pl.ResultChan <- Result{Text: "Sure! Send me the song or YouTube link!", State: c.downloadSong}
}

func (c *SeparateSongCommand) downloadSong(ctx context.Context, pl Payload) {
	res := Result{Text: "Downloading the song..."}
	pl.ResultChan <- res
	if pl.Media != nil {
		// c.blob.DownloadTelegramMedia(pl.UserID, pl.Media.FileID, pl.Media.FileURL)
	} else {
		// c.blob.DownloadYouTubeMedia(pl.UserID, pl.Command)
	}
}

func (c *SeparateSongCommand) queueProcessing(ctx context.Context, pl Payload) {
	pl.ResultChan <- Result{Text: "Ok! Loading and processing..."}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if c.queue.Lock() {
			defer c.queue.Unlock()

			var blobID int64
			// var err error

			// if pl.FileURL != "" {
			// 	blobID, err = c.blob.DownloadTelegramAudioIntoBlob(pl.FileURL)
			// 	if err != nil {
			// 		pl.ResultChan <- Result{
			// 			Text:  "Whoops, something is not right. Please try again later!",
			// 			State: c.queueProcessing,
			// 			Error: err,
			// 		}
			// 	}
			// } else {
			// 	blobID, err = c.blob.DownloadYouTubeAudioIntoBlob(pl.Command)
			// 	if err != nil {
			// 		pl.ResultChan <- Result{
			// 			Text:  "Sorry, download did not work. Can you try a different video link?",
			// 			State: c.startProcessing,
			// 			Error: err,
			// 		}
			// 	}
			// }

			// 2. Separate voice from music using python script
			// err = c.runPythonSeparation(ctx, blobID)
			// if err != nil {
			// 	pl.ResultChan <- Result{Text: err.Error(), State: c.startProcessing, Error: err}
			// 	return
			// }

			// 3. Send successful message
			res := Result{Text: "Song has been successfully processed!"}
			res.AddKeyboardButton("Get Music", commandf(c, "get_music", blobID))
			res.AddKeyboardButton("Get Voice", commandf(c, "get_voice", blobID))
			pl.ResultChan <- res

		} else {
			pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
		}
	}()

	wg.Wait()
}

func (c *SeparateSongCommand) processFile(ctx context.Context, pl Payload) {

}

func (c *SeparateSongCommand) runPythonSeparation(ctx context.Context, blobID int64) error {
	cmd := exec.CommandContext(
		ctx,
		"/home/dx/source/audio-separator/.venv/bin/python",
		"/home/dx/source/audio-separator/cli.py",
		"/home/dx/source/audio-separator/tracks/"+"smash.mp3",
		"--model_file_dir=/home/dx/source/audio-separator/models/",
		"--output_dir=/home/dx/source/audio-separator/tracks/",
		"--use_cuda",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
