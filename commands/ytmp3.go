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
	return "youtube to mp3"
}

func (c *YTMP3Command) Execute(ctx context.Context, pl Payload) {
	pl.ResultChan <- Result{Text: "Sure! Send me the YouTube link!", State: c.downloadSong}
}

func (c *YTMP3Command) downloadSong(ctx context.Context, pl Payload) {
	res := Result{Text: "🌐 Please wait..."}
	res.InlineMarkup.AddKeyboardButton("Downloading...", "-")
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	downloadedFile, err := utils.Download(ctx, pl.Command, "")
	if err != nil {
		res = Result{State: c.downloadSong, Error: err}
		if !errors.Is(err, context.Canceled) {
			res.Text = "Whoops, download failed, try again :c"
		}
		res.InlineMarkup.AddKeyboardRow()
		pl.ResultChan <- res
		return
	}

	res = Result{Text: fmt.Sprintf("📂 %s\n", _es(downloadedFile.Name))}
	res.InlineMarkup.AddKeyboardButton("Done!", "-")
	pl.ResultChan <- res

	pl.ResultChan <- Result{Audio: map[string]string{downloadedFile.Name: downloadedFile.Path}}
}
