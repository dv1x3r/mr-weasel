package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mr-weasel/utils"
)

type ExtractVoiceCommand struct {
	queue      *utils.Queue
	mode       string
	pathCLI    string
	pathModels string
	pathOutput string
}

func NewExtractVoiceCommand(queue *utils.Queue) *ExtractVoiceCommand {
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &ExtractVoiceCommand{
			queue:      queue,
			mode:       "CUDA",
			pathCLI:    filepath.Join(utils.GetExecutablePath(), "audio-separator", "bin", "audio-separator"),
			pathModels: filepath.Join(utils.GetExecutablePath(), "audio-separator", "models"),
			pathOutput: filepath.Join(utils.GetExecutablePath(), "audio-separator", "output"),
		}
	} else {
		return &ExtractVoiceCommand{
			queue:      queue,
			mode:       "CPU",
			pathCLI:    filepath.Join(utils.GetExecutablePath(), "audio-separator", "bin", "audio-separator"),
			pathModels: filepath.Join(utils.GetExecutablePath(), "audio-separator", "models"),
			pathOutput: filepath.Join(utils.GetExecutablePath(), "audio-separator", "output"),
		}
	}
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
		pl.ResultChan <- Result{Text: "Sure! Send me the song file or YouTube link!", State: c.downloadSong}
	}
}

func (c *ExtractVoiceCommand) downloadSong(ctx context.Context, pl Payload) {
	res := Result{Text: "🌐 Please wait..."}
	res.InlineMarkup.AddKeyboardButton("Downloading...", "-")
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	var downloadedFile utils.DownloadedFile
	var err error

	if pl.FileURL != "" {
		downloadedFile, err = utils.Download(ctx, pl.FileURL, pl.Command)
	} else {
		downloadedFile, err = utils.Download(ctx, pl.Command, "")
	}

	if err != nil {
		res = Result{State: c.downloadSong, Error: err}
		if !errors.Is(err, context.Canceled) {
			res.Text = "Whoops, download failed, try again :c"
		}
		res.InlineMarkup.AddKeyboardRow()
		pl.ResultChan <- res
		return
	}

	res = Result{Text: fmt.Sprintf("📂 %s\n", downloadedFile.Name)}
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("Start Processing %s", c.mode), commandf(c, cmdExtractVoiceStart, downloadedFile.ID))
	pl.ResultChan <- res
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

	model := "UVR-MDX-NET-Voc_FT" // best for vocal
	// model := "UVR-MDX-NET-Inst_HQ_3" // best for music

	var cmd *exec.Cmd

	switch c.mode {
	case "CUDA":
		cmd = exec.CommandContext(ctx, c.pathCLI, downloadedFile.Path,
			"--model_name", model,
			"--model_file_dir", c.pathModels,
			"--output_dir", c.pathOutput,
			"--output_format=MP3",
			"--log_level=DEBUG",
			"--use_cuda",
		)
	default:
		cmd = exec.CommandContext(ctx, c.pathCLI, downloadedFile.Path,
			"--model_name", model,
			"--model_file_dir", c.pathModels,
			"--output_dir", c.pathOutput,
			"--output_format=MP3",
			"--log_level=DEBUG",
		)
	}

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if err != nil {
		res = Result{}
		res.InlineMarkup.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, downloadedFile.ID))
		pl.ResultChan <- res
		if err.Error() != "signal: killed" {
			pl.ResultChan <- Result{Text: "Whoops, python script failed, try again :c", Error: err}
		}
		return
	}

	res = Result{}
	res.InlineMarkup.AddKeyboardButton("Done!", "-")
	pl.ResultChan <- res

	baseName := strings.TrimSuffix(filepath.Base(downloadedFile.Path), filepath.Ext(downloadedFile.Name))

	musicName := "Instrumental_" + downloadedFile.Name
	musicPath := filepath.Join(c.pathOutput, fmt.Sprintf("%s_(Instrumental)_%s.mp3", baseName, model))

	voiceName := "Vocals_" + downloadedFile.Name
	voicePath := filepath.Join(c.pathOutput, fmt.Sprintf("%s_(Vocals)_%s.mp3", baseName, model))

	pl.ResultChan <- Result{
		Audio: map[string]string{
			musicName: musicPath,
			voiceName: voicePath,
		},
	}
}
