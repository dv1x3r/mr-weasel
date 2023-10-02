package commands

import (
	"context"
	"os/exec"
	"path/filepath"

	st "mr-weasel/storage"
	"mr-weasel/utils"
)

type ChangeVoiceCommand struct {
	storage      *st.RvcStorage
	queue        *utils.Queue
	mode         string
	pathPython   string
	pathInferCLI string
	pathTrainCLI string
}

func NewChangeVoiceCommand(storage *st.RvcStorage, queue *utils.Queue) *ChangeVoiceCommand {
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &ChangeVoiceCommand{
			storage:      storage,
			queue:        queue,
			mode:         "CUDA",
			pathPython:   "/mnt/d/rvc-project/.venv/Scripts/python.exe",
			pathInferCLI: "/mnt/d/rvc-project/infer-cli.py",
			pathTrainCLI: "/mnt/d/rvc-project/train-cli.py",
		}
	} else {
		return &ChangeVoiceCommand{
			storage:      storage,
			queue:        queue,
			mode:         "CPU",
			pathPython:   filepath.Join(utils.GetExecutablePath(), "rvc-project", ".venv", "bin", "python"),
			pathInferCLI: filepath.Join(utils.GetExecutablePath(), "rvc-project", "infer-cli.py"),
			pathTrainCLI: filepath.Join(utils.GetExecutablePath(), "rvc-project", "train-cli.py"),
		}
	}
}

func (ChangeVoiceCommand) Prefix() string {
	return "/changevoice"
}

func (ChangeVoiceCommand) Description() string {
	return "train and change voices"
}

const (
	cmdChangeVoiceSelectModel = "select_model"
	cmdChangeVoiceSelectAudio = "select_audio"
	cmdChangeVoiceSetToneM12  = "set_tone_-12"
	cmdChangeVoiceSetToneM1   = "set_tone_-1"
	cmdChangeVoiceSetToneS0   = "set_tone_0"
	cmdChangeVoiceSetToneP1   = "set_tone_+1"
	cmdChangeVoiceSetToneP12  = "set_tone_+12"
	cmdChangeVoiceStartInfer  = "start_infer"
)

func (c *ChangeVoiceCommand) Execute(ctx context.Context, pl Payload) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	// case cmdExtractVoiceStart:
	// 	c.startProcessing(ctx, pl, strings.Join(args[1:], " "))
	default:
		c.newExperiment(ctx, pl)
	}
}

func (c *ChangeVoiceCommand) newExperiment(ctx context.Context, pl Payload) {
	// c.showMainMenu(ctx, pl, experiment)
	c.showMainMenu(ctx, pl)
}

func (c *ChangeVoiceCommand) showMainMenu(ctx context.Context, pl Payload) {
	res := Result{Text: "Experinemt info..."}
	res.InlineMarkup.AddKeyboardButton("Select Model", commandf(c, cmdChangeVoiceSelectModel))
	res.InlineMarkup.AddKeyboardButton("Select Audio", commandf(c, cmdChangeVoiceSelectAudio))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("-12 ♫", commandf(c, cmdChangeVoiceSetToneM12))
	res.InlineMarkup.AddKeyboardButton("-1 ♫", commandf(c, cmdChangeVoiceSetToneM1))
	res.InlineMarkup.AddKeyboardButton("0 ♫", commandf(c, cmdChangeVoiceSetToneS0))
	res.InlineMarkup.AddKeyboardButton("+1 ♫", commandf(c, cmdChangeVoiceSetToneP1))
	res.InlineMarkup.AddKeyboardButton("+12 ♫", commandf(c, cmdChangeVoiceSetToneP12))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Start Processing", commandf(c, cmdChangeVoiceStartInfer))
	pl.ResultChan <- res
}

// func (c *ExtractVoiceCommand) downloadSong(ctx context.Context, pl Payload) {
// 	res := Result{Text: "🌐 Please wait..."}
// 	res.AddKeyboardButton("Downloading...", "-")
// 	res.AddKeyboardRow()
// 	res.AddKeyboardButton("Cancel", cancelf(ctx))
// 	pl.ResultChan <- res

// 	var downloadedFile utils.DownloadedFile
// 	var err error

// 	if pl.FileURL != "" {
// 		downloadedFile, err = utils.Download(ctx, pl.FileURL, pl.Command)
// 	} else {
// 		downloadedFile, err = utils.Download(ctx, pl.Command, "")
// 	}

// 	if err != nil {
// 		res = Result{State: c.downloadSong, Error: err}
// 		if !errors.Is(err, context.Canceled) {
// 			res.Text = "Whoops, download failed, try again :c"
// 		}
// 		res.AddKeyboardRow()
// 		pl.ResultChan <- res
// 		return
// 	}

// 	res = Result{Text: fmt.Sprintf("📂 %s\n", downloadedFile.Name)}
// 	res.AddKeyboardButton(fmt.Sprintf("Start Processing %s", c.mode), commandf(c, cmdExtractVoiceStart, downloadedFile.UniqueID))
// 	pl.ResultChan <- res
// }

// func (c *ExtractVoiceCommand) startProcessing(ctx context.Context, pl Payload, uniqueID string) {
// 	res := Result{}
// 	res.AddKeyboardButton("Queued...", "-")
// 	res.AddKeyboardRow()
// 	res.AddKeyboardButton("Cancel", cancelf(ctx))
// 	pl.ResultChan <- res

// 	downloadedFile, err := utils.GetDownloadedFile(uniqueID)
// 	if err != nil {
// 		res := Result{}
// 		res.AddKeyboardButton("Error", "-")
// 		pl.ResultChan <- res
// 		pl.ResultChan <- Result{Text: "Whoops, file not available, try uploading again? :c", State: c.downloadSong, Error: err}
// 		return
// 	}

// 	if c.queue.Lock(ctx) {
// 		defer c.queue.Unlock()
// 		c.processFile(ctx, pl, downloadedFile)
// 	} else {
// 		res = Result{}
// 		res.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, uniqueID))
// 		pl.ResultChan <- res
// 		if !errors.Is(ctx.Err(), context.Canceled) {
// 			pl.ResultChan <- Result{Text: "There are too many queued jobs, please wait."}
// 		}
// 	}
// }

// func (c *ExtractVoiceCommand) processFile(ctx context.Context, pl Payload, downloadedFile utils.DownloadedFile) {
// 	res := Result{}
// 	res.AddKeyboardButton("Python goes brrr...", "-")
// 	res.AddKeyboardRow()
// 	res.AddKeyboardButton("Cancel", cancelf(ctx))
// 	pl.ResultChan <- res

// 	dir := utils.GetExecutablePath()
// 	model := "UVR-MDX-NET-Voc_FT" // best for vocal
// 	// model := "UVR-MDX-NET-Inst_HQ_3" // best for music

// 	execPath := filepath.Join(dir, "audio-separator", "bin", "audio-separator")
// 	modelsPath := filepath.Join(dir, "audio-separator", "models")
// 	outputPath := filepath.Join(dir, "audio-separator", "output")

// 	var cmd *exec.Cmd

// 	switch c.mode {
// 	case "CUDA":
// 		cmd = exec.CommandContext(ctx, execPath, downloadedFile.Path,
// 			"--log_level=DEBUG",
// 			"--model_name="+model,
// 			"--model_file_dir="+modelsPath,
// 			"--output_dir="+outputPath,
// 			"--output_format=MP3",
// 			"--use_cuda",
// 		)
// 	default:
// 		cmd = exec.CommandContext(ctx, execPath, downloadedFile.Path,
// 			"--log_level=DEBUG",
// 			"--model_name="+model,
// 			"--model_file_dir="+modelsPath,
// 			"--output_dir="+outputPath,
// 			"--output_format=MP3",
// 		)
// 	}

// 	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

// 	err := cmd.Run()
// 	if err != nil {
// 		res = Result{}
// 		res.AddKeyboardButton("Retry", commandf(c, cmdExtractVoiceStart, downloadedFile.UniqueID))
// 		pl.ResultChan <- res
// 		if err.Error() != "signal: killed" {
// 			pl.ResultChan <- Result{Text: "Whoops, python script failed, try again :c", Error: err}
// 		}
// 		return
// 	}

// 	os.Remove(downloadedFile.Path)

// 	res = Result{}
// 	res.AddKeyboardButton("Done!", "-")
// 	pl.ResultChan <- res

// 	baseName := strings.TrimSuffix(filepath.Base(downloadedFile.Path), filepath.Ext(downloadedFile.Name))

// 	musicName := "Instrumental_" + downloadedFile.Name
// 	musicPath := filepath.Join(outputPath, fmt.Sprintf("%s_(Instrumental)_%s.mp3", baseName, model))

// 	voiceName := "Vocals_" + downloadedFile.Name
// 	voicePath := filepath.Join(outputPath, fmt.Sprintf("%s_(Vocals)_%s.mp3", baseName, model))

// 	pl.ResultChan <- Result{
// 		Audio: map[string]string{
// 			musicName: musicPath,
// 			voiceName: voicePath,
// 		},
// 	}
// }
