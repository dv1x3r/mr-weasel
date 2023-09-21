package commands

import (
	"context"
	"errors"
	"fmt"

	"mr-weasel/utils"
)

type YTMP3Command struct {
}

func NewYTMP3Command() *YTMP3Command {
	return &YTMP3Command{}
}

func (YTMP3Command) Prefix() string {
	return "/ytmp3"
}

func (YTMP3Command) Description() string {
	return "download mp3 audio from YouTube"
}

func (c *YTMP3Command) Execute(ctx context.Context, pl Payload) {
	pl.ResultChan <- Result{Text: "Sure! Send me the YouTube link!", State: c.downloadSong}
}

func (c *YTMP3Command) downloadSong(ctx context.Context, pl Payload) {
	res := Result{Text: "🌐 Please wait..."}
	res.AddKeyboardButton("Downloading...", "-")
	res.AddKeyboardRow()
	res.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	downloadedFile, err := utils.Download(ctx, pl.Command, "")

	if err != nil {
		res = Result{State: c.downloadSong, Error: err}
		if !errors.Is(err, context.Canceled) {
			res.Text = "Whoops, download failed, try again :c"
		}
		res.AddKeyboardRow()
		pl.ResultChan <- res
		return
	}

	res = Result{Text: fmt.Sprintf("📂 %s\n", downloadedFile.Name)}
	res.AddKeyboardButton("Done!", "-")
	pl.ResultChan <- res

	pl.ResultChan <- Result{Audio: map[string]string{downloadedFile.Name: downloadedFile.Path}}
}
