package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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
	pathDatasets string
	pathWeights  string
	pathLogs     string
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
			pathDatasets: "/mnt/d/rvc-project/assets/datasets",
			pathWeights:  "/mnt/d/rvc-project/assets/weights",
			pathLogs:     "/mnt/d/rvc-project/logs",
		}
	} else {
		return &ChangeVoiceCommand{
			storage:      storage,
			queue:        queue,
			mode:         "CPU",
			pathPython:   filepath.Join(utils.GetExecutablePath(), "rvc-project", ".venv", "bin", "python"),
			pathInferCLI: filepath.Join(utils.GetExecutablePath(), "rvc-project", "infer-cli.py"),
			pathTrainCLI: filepath.Join(utils.GetExecutablePath(), "rvc-project", "train-cli.py"),
			pathDatasets: filepath.Join(utils.GetExecutablePath(), "rvc-project", "assets", "datasets"),
			pathWeights:  filepath.Join(utils.GetExecutablePath(), "rvc-project", "assets", "weights"),
			pathLogs:     filepath.Join(utils.GetExecutablePath(), "rvc-project", "logs"),
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
	cmdChangeVoiceExperimentGet = "experiment_get"
	cmdChangeVoiceModelGet      = "model_get"
	cmdChangeVoiceUploadAudio   = "upload_audio"
	cmdChangeVoiceEnableUVR     = "enable_uvr"
	cmdChangeVoiceDisableUVR    = "disable_uvr"
	cmdChangeVoiceSetModel      = "set_model"
	cmdChangeVoiceSetToneM12    = "set_tone_-12"
	cmdChangeVoiceSetToneM1     = "set_tone_-1"
	cmdChangeVoiceSetToneS0     = "set_tone_0"
	cmdChangeVoiceSetToneP1     = "set_tone_+1"
	cmdChangeVoiceSetToneP12    = "set_tone_+12"
	cmdChangeVoiceModelAdd      = "model_add"
	cmdChangeVoiceModelDelAsk   = "model_del"
	cmdChangeVoiceModelDelYes   = "model_del_yes"
	cmdChangeVoiceAccessDelYes  = "access_del_yes"
	cmdChangeVoiceAccessAdd     = "access_add"
	cmdChangeVoiceStart         = "start"
)

func (c *ChangeVoiceCommand) Execute(ctx context.Context, pl Payload) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdChangeVoiceExperimentGet:
		c.showExperimentDetails(ctx, pl, safeGetInt64(args, 1))
	case cmdChangeVoiceModelGet:
		c.showModelDetails(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdChangeVoiceUploadAudio:
		c.selectAudio(ctx, pl, safeGetInt64(args, 1))
	case cmdChangeVoiceEnableUVR:
		c.setExperimentSeparateUVR(ctx, pl, safeGetInt64(args, 1), true)
	case cmdChangeVoiceDisableUVR:
		c.setExperimentSeparateUVR(ctx, pl, safeGetInt64(args, 1), false)
	case cmdChangeVoiceSetModel:
		c.setExperimentModel(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdChangeVoiceSetToneM12:
		c.setExperimentTranspose(ctx, pl, safeGetInt64(args, 1), -12)
	case cmdChangeVoiceSetToneM1:
		c.setExperimentTranspose(ctx, pl, safeGetInt64(args, 1), -1)
	case cmdChangeVoiceSetToneS0:
		c.setExperimentTranspose(ctx, pl, safeGetInt64(args, 1), 0)
	case cmdChangeVoiceSetToneP1:
		c.setExperimentTranspose(ctx, pl, safeGetInt64(args, 1), 1)
	case cmdChangeVoiceSetToneP12:
		c.setExperimentTranspose(ctx, pl, safeGetInt64(args, 1), 12)
	case cmdChangeVoiceModelAdd:
		c.addModelStart(ctx, pl, safeGetInt64(args, 1))
	case cmdChangeVoiceAccessAdd:
		c.addAccessStart(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdChangeVoiceModelDelAsk:
		c.deleteModelAsk(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdChangeVoiceModelDelYes:
		c.deleteModelConfirm(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdChangeVoiceAccessDelYes:
		c.deleteAccessConfirm(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdChangeVoiceStart:
		c.startProcessing(ctx, pl, safeGetInt64(args, 1))
	default:
		c.newExperiment(ctx, pl)
	}
}

func (c *ChangeVoiceCommand) newExperiment(ctx context.Context, pl Payload) {
	experimentID, err := c.storage.InsertNewExperimentIntoDB(ctx, pl.UserID)
	if err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
	} else {
		c.showExperimentDetails(ctx, pl, experimentID)
	}
}

func (c *ChangeVoiceCommand) formatExperimentDetails(experiment st.RvcExperimentDetails) string {
	str := ""
	if experiment.ModelName.Valid {
		str += fmt.Sprintf("🗣️ <b>Model:</b> %s\n", _es(experiment.ModelName.String))
	} else {
		str += fmt.Sprintf("🗣️ <b>Model:</b> 🚫 Not Selected\n")
	}
	if experiment.AudioSourceID.Valid {
		audioFile, _ := utils.GetDownloadedFile(experiment.AudioSourceID.String)
		if experiment.SeparateUVR.Bool {
			str += fmt.Sprintf("🎺 <b>Audio with music:</b> %s\n", _es(audioFile.Name))
		} else {
			str += fmt.Sprintf("🎤 <b>Audio acapella:</b> %s\n", _es(audioFile.Name))
		}
	} else {
		str += fmt.Sprintf("🎧 <b>Audio:</b> 🚫 Not Selected\n")
	}
	str += fmt.Sprintf("🎼 <b>Transpose:</b> %+d semitones\n", experiment.Transpose.Int64)
	return str
}

func (c *ChangeVoiceCommand) showExperimentDetails(ctx context.Context, pl Payload, experimentID int64) {
	res := Result{ClearState: true}
	experiment, err := c.storage.GetExperimentDetailsFromDB(ctx, pl.UserID, experimentID)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "Experiment not found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatExperimentDetails(experiment)
		res.InlineMarkup.AddKeyboardButton("Select Model", commandf(c, cmdChangeVoiceModelGet, experimentID))
		res.InlineMarkup.AddKeyboardButton("Select Audio", commandf(c, cmdChangeVoiceUploadAudio, experimentID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("-12 ♫", commandf(c, cmdChangeVoiceSetToneM12, experimentID))
		res.InlineMarkup.AddKeyboardButton("-1 ♫", commandf(c, cmdChangeVoiceSetToneM1, experimentID))
		res.InlineMarkup.AddKeyboardButton("0 ♫", commandf(c, cmdChangeVoiceSetToneS0, experimentID))
		res.InlineMarkup.AddKeyboardButton("+1 ♫", commandf(c, cmdChangeVoiceSetToneP1, experimentID))
		res.InlineMarkup.AddKeyboardButton("+12 ♫", commandf(c, cmdChangeVoiceSetToneP12, experimentID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Start Processing", commandf(c, cmdChangeVoiceStart, experimentID))
	}
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) formatModelDetails(model st.RvcModelDetails) string {
	str := fmt.Sprintf("🗣️ <b>Model:</b> %s\n", _es(model.Name))
	if model.IsOwner {
		str += fmt.Sprintf("🔑 <b>Access:</b> Full access\n")
		str += fmt.Sprintf("🌐 <b>Shared with:</b> %d contacts\n", model.Shares)
	} else {
		str += fmt.Sprintf("🔑 <b>Access:</b> Shared with you\n")
	}
	return str
}

func (c *ChangeVoiceCommand) showModelDetails(ctx context.Context, pl Payload, experimentID int64, offset int64) {
	res := Result{}
	model, err := c.storage.GetModelFromDB(ctx, pl.UserID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No models found."
		res.InlineMarkup.AddKeyboardButton("« New »", commandf(c, cmdChangeVoiceModelAdd, experimentID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("« Back", commandf(c, cmdChangeVoiceExperimentGet, experimentID))
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
		res.InlineMarkup.AddKeyboardButton("« Back", commandf(c, cmdChangeVoiceExperimentGet, experimentID))
	} else {
		res.Text = c.formatModelDetails(model)
		res.InlineMarkup.AddKeyboardPagination(offset, model.CountRows, commandf(c, cmdChangeVoiceModelGet, experimentID))
		res.InlineMarkup.AddKeyboardRow()
		if model.IsOwner {
			res.InlineMarkup.AddKeyboardButton("Delete", commandf(c, cmdChangeVoiceModelDelAsk, experimentID, model.ID))
			if pl.IsPrivate {
				res.InlineMarkup.AddKeyboardButton("Share", commandf(c, cmdChangeVoiceAccessAdd, experimentID, model.ID))
			}
			res.InlineMarkup.AddKeyboardRow()
		}
		res.InlineMarkup.AddKeyboardButton("« Back", commandf(c, cmdChangeVoiceExperimentGet, experimentID))
		res.InlineMarkup.AddKeyboardButton("« New »", commandf(c, cmdChangeVoiceModelAdd, experimentID))
		res.InlineMarkup.AddKeyboardButton("Select", commandf(c, cmdChangeVoiceSetModel, experimentID, model.ID))
	}
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) selectAudio(ctx context.Context, pl Payload, experimentID int64) {
	res := Result{
		Text:  "Send me the a YouTube link, a song file, or record a new voice message!",
		State: func(ctx context.Context, pl Payload) { c.setExperimentAudioSource(ctx, pl, experimentID) },
	}
	res.InlineMarkup.AddKeyboardButton("« Back", commandf(c, cmdChangeVoiceExperimentGet, experimentID))
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) setExperimentAudioSource(ctx context.Context, pl Payload, experimentID int64) {
	res := Result{Text: "🌐 Please wait..."}
	res.InlineMarkup.AddKeyboardButton("Downloading...", "-")
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Cancel", cancelf(ctx))
	pl.ResultChan <- res

	downloadedFile, err := utils.Download(ctx, pl.FileURL, pl.Command)
	if errors.Is(err, context.Canceled) {
		res = Result{
			Text:  "Download cancelled, you can send another audio.",
			State: func(ctx context.Context, pl Payload) { c.setExperimentAudioSource(ctx, pl, experimentID) },
			Error: err,
		}
		res.InlineMarkup.AddKeyboardRow()
		pl.ResultChan <- res
	} else if err != nil {
		res = Result{
			Text:  "Whoops, download failed, try again :c",
			State: func(ctx context.Context, pl Payload) { c.setExperimentAudioSource(ctx, pl, experimentID) },
			Error: err,
		}
		res.InlineMarkup.AddKeyboardRow()
		pl.ResultChan <- res
	} else {
		err := c.storage.SetExperimentAudioSourceInDB(ctx, pl.UserID, experimentID, downloadedFile.ID)
		if err != nil {
			pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		} else {
			res = Result{Text: "Does it contain music, or voice only?"}
			res.InlineMarkup.AddKeyboardButton("Music and Voice", commandf(c, cmdChangeVoiceEnableUVR, experimentID))
			res.InlineMarkup.AddKeyboardRow()
			res.InlineMarkup.AddKeyboardButton("Voice only", commandf(c, cmdChangeVoiceDisableUVR, experimentID))
			pl.ResultChan <- res
		}
	}
}

func (c *ChangeVoiceCommand) setExperimentSeparateUVR(ctx context.Context, pl Payload, experimentID int64, value bool) {
	if err := c.storage.SetExperimentSeparateUVRInDB(ctx, pl.UserID, experimentID, value); err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
	} else {
		c.showExperimentDetails(ctx, pl, experimentID)
	}
}

func (c *ChangeVoiceCommand) setExperimentModel(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	if err := c.storage.SetExperimentModelInDB(ctx, pl.UserID, experimentID, modelID); err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
	} else {
		c.showExperimentDetails(ctx, pl, experimentID)
	}
}

func (c *ChangeVoiceCommand) setExperimentTranspose(ctx context.Context, pl Payload, experimentID int64, delta int64) {
	experiment, err := c.storage.GetExperimentDetailsFromDB(ctx, pl.UserID, experimentID)
	if err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
	}

	var newValue int64

	if delta == 0 {
		newValue = 0
	} else {
		newValue = experiment.Transpose.Int64 + delta
		if newValue > 24 {
			newValue = 24
		} else if newValue < -24 {
			newValue = -24
		}
	}

	if experiment.Transpose.Int64 == newValue {
		return
	}

	if err := c.storage.SetExperimentTransposeInDB(ctx, pl.UserID, experimentID, newValue); err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
	} else {
		c.showExperimentDetails(ctx, pl, experimentID)
	}
}

func (c *ChangeVoiceCommand) addModelStart(ctx context.Context, pl Payload, experimentID int64) {
	pl.ResultChan <- Result{
		Text:  "Let's create a new voice model. How should we name it?",
		State: func(ctx context.Context, pl Payload) { c.addModelNameAndSave(ctx, pl, experimentID) },
	}
}

func (c *ChangeVoiceCommand) addModelNameAndSave(ctx context.Context, pl Payload, experimentID int64) {
	modelID, err := c.storage.InsertNewModelIntoDB(ctx, pl.UserID, pl.Command)
	if err != nil {
		res := Result{Text: "There is something wrong, please try again.", Error: err}
		res.InlineMarkup.AddKeyboardButton("« Back to my models", commandf(c, cmdChangeVoiceModelGet, experimentID))
		pl.ResultChan <- res
	} else {
		msg := fmt.Sprintf("Ok! Now send me voice samples with audio files or voice messages.\n\n")
		msg += fmt.Sprintf("They should be without any background music, and with as little noise as possible. ")
		msg += fmt.Sprintf("For simple models even 10 seconds of audio is enough. For better quality we need 40-60 seconds in total.\n\n")
		msg += fmt.Sprintf("When you are ready, just send /done command!")
		pl.ResultChan <- Result{
			Text:  msg,
			State: func(ctx context.Context, pl Payload) { c.addModelDatasetFile(ctx, pl, experimentID, modelID) },
		}
	}
}

func (c *ChangeVoiceCommand) addModelDatasetFile(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	if pl.Command == "/done" {
		c.setExperimentModel(ctx, pl, experimentID, modelID)
		return
	}

	downloadedFile, err := utils.Download(ctx, pl.FileURL, pl.Command)
	if err != nil {
		pl.ResultChan <- Result{
			Text:  "Whoops, download failed, try again :c",
			State: func(ctx context.Context, pl Payload) { c.addModelDatasetFile(ctx, pl, experimentID, modelID) },
			Error: err,
		}
	} else {
		// Move downloaded file to the datasets directory
		os.MkdirAll(filepath.Join(c.pathDatasets, fmt.Sprint(modelID)), os.ModePerm)
		os.Rename(downloadedFile.Path, filepath.Join(c.pathDatasets, fmt.Sprint(modelID), filepath.Base(downloadedFile.Path)))
		pl.ResultChan <- Result{
			Text:  _es(fmt.Sprintf("<b>%s</b> has been imported!", downloadedFile.Name)),
			State: func(ctx context.Context, pl Payload) { c.addModelDatasetFile(ctx, pl, experimentID, modelID) },
		}
	}
}

func (c *ChangeVoiceCommand) addAccessStart(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	res := Result{
		Text:  "Select the contact with whom you would like to share the selected model. Use the button below the keyboard.",
		State: func(ctx context.Context, pl Payload) { c.addAccessUser(ctx, pl, experimentID, modelID) },
	}
	res.ReplyMarkup.AddRequestUserButton()
	res.ReplyMarkup.AddKeyboardRow()
	res.ReplyMarkup.AddButton("Close")
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) addAccessUser(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	if pl.Command == "Close" {
		res := Result{Text: "Done!"}
		res.RemoveMarkup.RemoveDefault()
		pl.ResultChan <- res
		return
	}

	accessUserID, err := strconv.Atoi(pl.Command)
	if err != nil {
		pl.ResultChan <- Result{
			Text:  "You need to select the contact with button below.",
			State: func(ctx context.Context, pl Payload) { c.addAccessUser(ctx, pl, experimentID, modelID) },
			Error: err,
		}
		return
	}

	_, err = c.storage.InsertNewAccessIntoDB(ctx, pl.UserID, modelID, int64(accessUserID))
	if err != nil {
		pl.ResultChan <- Result{
			Text:  "There is something wrong, please try again.",
			State: func(ctx context.Context, pl Payload) { c.addAccessUser(ctx, pl, experimentID, modelID) },
			Error: err,
		}
	}
}

func (c *ChangeVoiceCommand) deleteModelAsk(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	res := Result{Text: "Are you sure you want to delete the selected model, or reset access?"}
	res.InlineMarkup.AddKeyboardButton("Yes, delete the model", commandf(c, cmdChangeVoiceModelDelYes, experimentID, modelID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Reset access", commandf(c, cmdChangeVoiceAccessDelYes, experimentID, modelID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Nope, nevermind", commandf(c, cmdChangeVoiceModelGet, experimentID))
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) deleteModelConfirm(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	res := Result{}
	affected, err := c.storage.DeleteModelFromDB(ctx, pl.UserID, modelID)
	if err != nil || affected != 1 {
		res.Text, res.Error = "Model not found.", err
	} else {
		res.Text = "Model has been successfully deleted!"
		os.RemoveAll(filepath.Join(c.pathDatasets, fmt.Sprint(modelID)))          // delete datasets folder
		os.RemoveAll(filepath.Join(c.pathLogs, fmt.Sprint(modelID)))              // delete logs folder
		os.Remove(filepath.Join(c.pathWeights, fmt.Sprintf("%d.pth", modelID)))   // delete model weights
		os.Remove(filepath.Join(c.pathWeights, fmt.Sprintf("%d.index", modelID))) // delete model index
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my models", commandf(c, cmdChangeVoiceModelGet, experimentID))
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) deleteAccessConfirm(ctx context.Context, pl Payload, experimentID int64, modelID int64) {
	res := Result{}
	_, err := c.storage.DeleteAccessFromDB(ctx, pl.UserID, modelID)
	if err != nil {
		res.Text, res.Error = "Model not found.", err
	} else {
		res.Text = "Permisions has been revoked, now only you can access this model."
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my models", commandf(c, cmdChangeVoiceModelGet, experimentID))
	pl.ResultChan <- res
}

func (c *ChangeVoiceCommand) startProcessing(ctx context.Context, pl Payload, experimentID int64) {
}

// func (c *ChangeVoiceCommand) testUser(ctx context.Context, pl Payload) {
// res := Result{Text: "Press button below to select new model user. /skip", State: c.testUser}
// res.ReplyMarkup.AddRequestUserButton()
// pl.ResultChan <- res
// 	if pl.Command == "/skip" {
// 		res := Result{Text: "skip, keyboard closed"}
// 		res.RemoveMarkup.RemoveDefault()
// 		pl.ResultChan <- res
// 	} else {
// 		res := Result{Text: fmt.Sprintf("%s selected, keyboard closed", pl.Command)}
// 		res.RemoveMarkup.RemoveDefault()
// 		pl.ResultChan <- res
// 	}
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
