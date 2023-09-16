package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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

const (
	cmdSeparateSongStart = "start"
	cmdSeparateSongMusic = "music"
	cmdSeparateSongVoice = "voice"
)

func (c *SeparateSongCommand) Execute(ctx context.Context, pl Payload) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdSeparateSongStart:
		c.startProcessing(ctx, pl, safeGetInt64(args, 1))
	case cmdSeparateSongMusic:
	// c.getMusic(ctx, pl, safeGetInt64(args, 1))
	case cmdSeparateSongVoice:
	// c.getVoice(ctx, pl, safeGetInt64(args, 1))
	default:
		pl.ResultChan <- Result{Text: "Sure! Send me the song or YouTube link!", State: c.downloadSong}
	}
}

func (c *SeparateSongCommand) downloadSong(ctx context.Context, pl Payload) {
	var blob utils.BlobBase
	var err error

	if pl.BlobPayload != nil {
		res := Result{Text: "📂 " + pl.BlobPayload.FileName}
		res.AddKeyboardButton("Downloading...", "-")
		pl.ResultChan <- res

		blob, err = c.blob.DownloadBlob(ctx, pl.UserID, pl.BlobPayload)
		if err != nil {
			res = Result{
				Text:  "Whoops, download failed, try again :c",
				State: c.downloadSong,
				Error: err,
			}
			res.AddKeyboardRow()
			pl.ResultChan <- res
			return
		}

	} else {
		res := Result{Text: "🌐 Please wait..."}
		res.AddKeyboardButton("Downloading...", "-")
		pl.ResultChan <- res

		blob, err = c.blob.DownloadYouTube(ctx, pl.UserID, pl.Command)
		if err != nil {
			res = Result{
				Text:  "Whoops, please try another link :c",
				State: c.downloadSong,
				Error: err,
			}
			res.AddKeyboardRow()
			pl.ResultChan <- res
			return
		}
	}

	res := Result{Text: fmt.Sprintf("📂 %s\n", blob.OriginalName)}
	res.AddKeyboardButton("Start Processing", commandf(c, cmdSeparateSongStart, blob.ID))
	pl.ResultChan <- res
}

func (c *SeparateSongCommand) startProcessing(ctx context.Context, pl Payload, blobID int64) {
	if c.queue.Lock() {
		defer c.queue.Unlock()
		c.processFile(ctx, pl, blobID)
	} else {
		pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
	}
}

func (c *SeparateSongCommand) processFile(ctx context.Context, pl Payload, blobID int64) {
	blob, err := c.blob.GetBlobFromDB(ctx, pl.UserID, blobID)
	if err != nil {
		pl.ResultChan <- Result{
			Text:  "File not found, can you please forward it to me?",
			State: c.downloadSong,
		}
		return
	}

	res := Result{}
	res.AddKeyboardButton("Python goes brrr...", "-")
	res.AddKeyboardRow()
	res.AddKeyboardButton("Cancel", commandf(c, "cancel"))
	pl.ResultChan <- res

	cmd := exec.CommandContext(
		ctx,
		"/home/dx/source/audio-separator/.venv/bin/python",
		"/home/dx/source/audio-separator/cli.py",
		blob.GetAbsolutePath(),
		"--model_file_dir=/home/dx/source/audio-separator/models/",
		"--output_dir=/home/dx/source/audio-separator/output/",
		"--use_cuda",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err = cmd.Run()
	if err != nil {
		res := Result{}
		res.AddKeyboardButton("Retry", commandf(c, cmdSeparateSongStart, blob.ID))
		pl.ResultChan <- res
		pl.ResultChan <- Result{Text: "Whoops, python script failed, try again :c", Error: err}
		return
	}

	res = Result{}
	res.AddKeyboardButton("Uploading results...", "-")
	res.AddKeyboardRow()
	res.AddKeyboardButton("Cancel", cmdCancel)
	pl.ResultChan <- res

	// TODO: upload here
	musicBlobID := 0
	voiceBlobID := 0

	res = Result{}
	res.AddKeyboardButton("Get Music", commandf(c, "music", musicBlobID))
	res.AddKeyboardButton("Get Voice", commandf(c, "voice", voiceBlobID))
	pl.ResultChan <- res
	pl.ResultChan <- Result{Text: "Song has been successfully processed!"}
}
