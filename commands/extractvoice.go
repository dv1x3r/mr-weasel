package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mr-weasel/utils"
)

type ExtractVoiceCommand struct {
	queue     *utils.Queue
	separator *utils.AudioSeparator
}

func NewExtractVoiceCommand(queue *utils.Queue, separator *utils.AudioSeparator) *ExtractVoiceCommand {
	return &ExtractVoiceCommand{queue: queue, separator: separator}
}

func (ExtractVoiceCommand) Prefix() string {
	return "/extractvoice"
}

func (ExtractVoiceCommand) Description() string {
	return "separate voice and music"
}

const (
	cmdExtractVoiceStart = "start"
)

func (c *ExtractVoiceCommand) Execute(ctx context.Context, pl Payload) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdExtractVoiceStart:
		c.startProcessing(ctx, pl, strings.Join(args[1:], " "))
	default:
		pl.ResultChan <- Result{Text: "Sure! Send me a YouTube link or song file!", State: c.downloadSong}
	}
}

func (c *ExtractVoiceCommand) downloadSong(ctx context.Context, pl Payload) {
	res := Result{Text: "🌐 Please wait..."}
	res.InlineMarkup.AddKeyboardButton("Downloading...", "-")
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	downloadedFile, err := utils.Download(ctx, pl.FileURL, pl.Command)
	if errors.Is(err, context.Canceled) {
		res = Result{Text: "Download cancelled, you can send another song.", State: c.downloadSong, Error: err}
		res.InlineMarkup.AddKeyboardRow()
		pl.ResultChan <- res
	} else if err != nil {
		res = Result{Text: "Whoops, download failed, try again :c", State: c.downloadSong, Error: err}
		res.InlineMarkup.AddKeyboardRow()
		pl.ResultChan <- res
	} else {
		res = Result{Text: fmt.Sprintf("📂 %s\n", _es(downloadedFile.Name))}
		res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("Start Processing %s", c.separator.Mode), commandf(c, cmdExtractVoiceStart, downloadedFile.ID))
		pl.ResultChan <- res
	}
}

func (c *ExtractVoiceCommand) startProcessing(ctx context.Context, pl Payload, uniqueID string) {
	res := Result{}
	res.InlineMarkup.AddKeyboardButton("Queued...", "-")
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	downloadedFile, err := utils.GetDownloadedFile(uniqueID)
	if err != nil {
		res := Result{}
		res.InlineMarkup.AddKeyboardButton("Error", "-")
		pl.ResultChan <- res
		pl.ResultChan <- Result{Text: "Whoops, file not available, try uploading again? :c", State: c.downloadSong, Error: err}
		return
	}

	if c.queue.Lock(ctx) {
		defer c.queue.Unlock()
		c.processFile(ctx, pl, downloadedFile)
	} else {
		res = Result{}
		res.InlineMarkup.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, uniqueID))
		pl.ResultChan <- res
		if !errors.Is(ctx.Err(), context.Canceled) {
			pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
		}
	}
}

func (c *ExtractVoiceCommand) processFile(ctx context.Context, pl Payload, downloadedFile utils.DownloadedFile) {
	res := Result{}
	res.InlineMarkup.AddKeyboardButton("Python goes brrr...", "-")
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	resFiles, err := c.separator.Run(ctx, downloadedFile)
	if errors.Is(err, context.Canceled) {
		res = Result{}
		res.InlineMarkup.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, downloadedFile.ID))
		pl.ResultChan <- res
		return
	} else if err != nil {
		res = Result{}
		res.InlineMarkup.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, downloadedFile.ID))
		pl.ResultChan <- res
		pl.ResultChan <- Result{Text: "Whoops, python script failed, try again :c", Error: err}
	}

	res = Result{}
	res.InlineMarkup.AddKeyboardButton("Done!", "-")
	pl.ResultChan <- res

	pl.ResultChan <- Result{
		Audio: map[string]string{
			resFiles.MusicName: resFiles.MusicPath,
			resFiles.VoiceName: resFiles.VoicePath,
		},
	}
}
