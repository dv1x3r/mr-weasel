package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mr-weasel/utils"
)

type ExtractVoiceCommand struct {
	queue *utils.Queue
}

func NewExtractVoiceCommand(queue *utils.Queue) *ExtractVoiceCommand {
	return &ExtractVoiceCommand{queue: queue}
}

func (ExtractVoiceCommand) Prefix() string {
	return "/extractvoice"
}

func (ExtractVoiceCommand) Description() string {
	return "separate voice from music"
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
		pl.ResultChan <- Result{Text: "Sure! Send me the song file or YouTube link!", State: c.downloadSong}
	}
}

func (c *ExtractVoiceCommand) downloadSong(ctx context.Context, pl Payload) {
	res := Result{Text: "🌐 Please wait..."}
	res.AddKeyboardButton("Downloading...", "-")
	res.AddKeyboardRow()
	res.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	var filePath string
	var err error

	if pl.FileURL != "" {
		filePath, err = utils.Download(ctx, pl.FileURL, pl.Command)
	} else {
		filePath, err = utils.Download(ctx, pl.Command, "")
	}

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

	fileBase := filepath.Base(filePath)
	originalName := fileBase

	split := strings.SplitN(fileBase, ".", 2)
	if len(split) == 2 {
		originalName = split[1]
	}

	res = Result{Text: fmt.Sprintf("📂 %s\n", originalName)}
	res.AddKeyboardButton("Start Processing", commandf(c, cmdExtractVoiceStart, fileBase))
	pl.ResultChan <- res
}

func (c *ExtractVoiceCommand) startProcessing(ctx context.Context, pl Payload, fileBase string) {
	res := Result{}
	res.AddKeyboardButton("Queued...", "-")
	res.AddKeyboardRow()
	res.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	if c.queue.Lock(ctx) {
		defer c.queue.Unlock()
		c.processFile(ctx, pl, fileBase)
	} else {
		res := Result{}
		res.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, fileBase))
		pl.ResultChan <- res
		pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
	}
}

func (c *ExtractVoiceCommand) processFile(ctx context.Context, pl Payload, fileBase string) {
	res := Result{}
	res.AddKeyboardButton("Python goes brrr...", "-")
	res.AddKeyboardRow()
	res.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	fileAbsolutePath, _ := filepath.Abs(filepath.Join(utils.DownloadDir, fileBase))

	model := "UVR-MDX-NET-Voc_FT" // best for vocal
	// model := "UVR-MDX-NET-Inst_HQ_3" // best for music

	cmd := exec.CommandContext(
		ctx,
		"/home/dx/source/audio-separator/.venv/bin/python",
		"/home/dx/source/audio-separator/cli.py",
		fileAbsolutePath,
		"--model_name="+model,
		"--model_file_dir=/home/dx/source/audio-separator/models/",
		"--output_dir=/home/dx/source/audio-separator/output/",
		"--use_cuda",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if err != nil {
		res := Result{}
		res.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, fileBase))
		pl.ResultChan <- res
		pl.ResultChan <- Result{Text: "Whoops, python script failed, try again :c", Error: err}
		return
	}

	os.Remove(fileAbsolutePath)

	res = Result{}
	res.AddKeyboardButton("Done!", "-")
	pl.ResultChan <- res

	originalName := fileBase

	split := strings.SplitN(fileBase, ".", 2)
	if len(split) == 2 {
		originalName = split[1]
	}

	musicName := "Instrumental_" + originalName
	musicPath := filepath.Join(
		"/home/dx/source/audio-separator/output/",
		fmt.Sprintf("%s_%s_%s.mp3", strings.TrimSuffix(fileBase, filepath.Ext(fileBase)), "(Instrumental)", model),
	)

	voiceName := "Vocals_" + originalName
	voicePath := filepath.Join(
		"/home/dx/source/audio-separator/output/",
		fmt.Sprintf("%s_%s_%s.mp3", strings.TrimSuffix(fileBase, filepath.Ext(fileBase)), "(Vocals)", model),
	)

	pl.ResultChan <- Result{
		Audio: map[string]string{
			musicName: musicPath,
			voiceName: voicePath,
		},
	}
}
