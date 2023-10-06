package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"mr-weasel/storage"
)

type VoiceChanger struct {
	Mode         string
	PathPython   string
	PathInferCLI string
	PathTrainCLI string
	PathDatasets string
	PathWeights  string
	PathLogs     string
	PathOutput   string
}

type VoiceChangerResult struct {
	Name string
	Path string
}

func NewVoiceChanger() *VoiceChanger {
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &VoiceChanger{
			Mode:         "CUDA",
			PathPython:   "/mnt/d/rvc-project/.venv/Scripts/python.exe",
			PathInferCLI: "D:\\rvc-project\\infer-cli.py",
			PathTrainCLI: "D:\\rvc-project\\train-cli.py",
			PathDatasets: "/mnt/d/rvc-project/assets/datasets",
			PathWeights:  "/mnt/d/rvc-project/assets/weights",
			PathLogs:     "/mnt/d/rvc-project/logs",
			PathOutput:   "/mnt/d/rvc-project/TEMP",
		}
	} else {
		return &VoiceChanger{
			Mode:         "CPU",
			PathPython:   filepath.Join(GetExecutablePath(), "rvc-project", ".venv", "bin", "python"),
			PathInferCLI: filepath.Join(GetExecutablePath(), "rvc-project", "infer-cli.py"),
			PathTrainCLI: filepath.Join(GetExecutablePath(), "rvc-project", "train-cli.py"),
			PathDatasets: filepath.Join(GetExecutablePath(), "rvc-project", "assets", "datasets"),
			PathWeights:  filepath.Join(GetExecutablePath(), "rvc-project", "assets", "weights"),
			PathLogs:     filepath.Join(GetExecutablePath(), "rvc-project", "logs"),
			PathOutput:   filepath.Join(GetExecutablePath(), "rvc-project", "TEMP"),
		}
	}
}

func (vc *VoiceChanger) DeleteAll(modelID int64) {
	os.RemoveAll(filepath.Join(vc.PathDatasets, fmt.Sprint(modelID)))          // delete datasets folder
	os.RemoveAll(filepath.Join(vc.PathLogs, fmt.Sprint(modelID)))              // delete logs folder
	os.Remove(filepath.Join(vc.PathWeights, fmt.Sprintf("%d.pth", modelID)))   // delete model weights
	os.Remove(filepath.Join(vc.PathWeights, fmt.Sprintf("%d.index", modelID))) // delete model index
}

func (vc *VoiceChanger) IsTrained(modelID int64) bool {
	pthPath := filepath.Join(filepath.Join(vc.PathWeights, fmt.Sprintf("%d.pth", modelID)))
	indexPath := filepath.Join(vc.PathWeights, fmt.Sprintf("%d.index", modelID))
	_, err1 := os.Stat(pthPath)
	_, err2 := os.Stat(indexPath)
	return errors.Join(err1, err2) == nil
}

func (vc *VoiceChanger) RunTrain(ctx context.Context, experiment storage.RvcExperimentDetails) error {
	modelFolder := fmt.Sprint(experiment.ModelID.Int64)

	var cmd *exec.Cmd
	switch vc.Mode {
	case "CUDA":
		cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathTrainCLI,
			"--name", modelFolder,
			"--dataset", filepath.Join("assets", "datasets", modelFolder),
			"--version", "v2",
			"--sample_rate", "40k",
			"--method", "rmvpe_gpu",
			"--gpu_rmvpe", "0-0",
			"--gpu", "0",
			"--batch_size", "8",
			"--total_epoch", "150",
			"--save_epoch", "10",
			"--save_latest", "1",
			"--cache_gpu", "0",
			"--save_every_weights", "0",
		)
	default:
		cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathTrainCLI,
			"--name", modelFolder,
			"--dataset", filepath.Join("assets", "datasets", modelFolder),
			"--version", "v2",
			"--sample_rate", "40k",
			"--method", "rmvpe",
			"--gpu_rmvpe", "-",
			"--gpu", "",
			"--batch_size", "1",
			"--total_epoch", "10",
			"--save_epoch", "2",
			"--save_latest", "1",
			"--cache_gpu", "0",
			"--save_every_weights", "0",
		)
	}

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if err != nil && err.Error() == "signal: killed" {
		return context.Canceled
	} else if err != nil {
		return fmt.Errorf("%w: %s", err, cmd.Stderr)
	}

	// move index to the weights folder, and remove logs
	MoveCrossDevice(
		filepath.Join(vc.PathLogs, modelFolder, "added.index"),
		filepath.Join(vc.PathWeights, fmt.Sprintf("%s.index", modelFolder)),
	)

	// TODO: clear training data after some tests
	// os.RemoveAll(filepath.Join(vc.PathLogs, modelName))
	// os.RemoveAll(filepath.Join(vc.PathDatasets, modelName))

	return nil
}

func (vc *VoiceChanger) RunInfer(ctx context.Context, experiment storage.RvcExperimentDetails, voicePath string) (VoiceChangerResult, error) {
	audioFile, _ := GetDownloadedFile(experiment.Audio.String)
	modelFolder := fmt.Sprint(experiment.ModelID.Int64)
	inputName := regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(filepath.Base(voicePath), "")

	baseName := strings.TrimSuffix(filepath.Base(audioFile.Name), filepath.Ext(audioFile.Name))
	outputNameWav := fmt.Sprintf("%s.%s.wav", experiment.ModelName.String, regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(baseName, ""))
	outputNameMp3 := fmt.Sprintf("%s.%s.mp3", experiment.ModelName.String, baseName)

	CopyCrossDevice(voicePath, filepath.Join(vc.PathOutput, inputName))
	defer os.Remove(filepath.Join(vc.PathOutput, inputName))
	defer os.Remove(filepath.Join(vc.PathOutput, outputNameWav))

	var cmd *exec.Cmd
	switch vc.Mode {
	case "CUDA":
		cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathInferCLI,
			"--input", filepath.Join("TEMP", inputName),
			"--output", filepath.Join("TEMP", outputNameWav),
			"--model", fmt.Sprintf("%s.pth", modelFolder),
			"--index", filepath.Join("assets", "weights", fmt.Sprintf("%s.index", modelFolder)),
			"--method", "rmvpe",
			"--transpose", fmt.Sprint(experiment.Transpose.Int64),
			// "--ratio", "0.75",
			// "--filter", "3",
			// "--resample", "0",
			// "--rms", "0.25",
			// "--protect", "0.33",
		)
	default:
		cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathInferCLI,
			"--input", filepath.Join("TEMP", inputName),
			"--output", filepath.Join("TEMP", outputNameWav),
			"--model", fmt.Sprintf("%s.pth", modelFolder),
			"--index", filepath.Join("assets", "weights", fmt.Sprintf("%s.index", modelFolder)),
			"--method", "pm",
			"--transpose", fmt.Sprint(experiment.Transpose.Int64),
			// "--ratio", "0.75",
			// "--filter", "3",
			// "--resample", "0",
			// "--rms", "0.25",
			// "--protect", "0.33",
		)
	}

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if err != nil && err.Error() == "signal: killed" {
		return VoiceChangerResult{}, context.Canceled
	} else if err != nil {
		return VoiceChangerResult{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
	}

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return VoiceChangerResult{}, err
	}

	cmd = exec.CommandContext(ctx, ffmpeg,
		"-i", filepath.Join(vc.PathOutput, outputNameWav),
		"-b:a", "320k",
		"-y",
		filepath.Join(vc.PathOutput, outputNameMp3),
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err = cmd.Run()
	if err != nil && err.Error() == "signal: killed" {
		return VoiceChangerResult{}, context.Canceled
	} else if err != nil {
		return VoiceChangerResult{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
	}

	res := VoiceChangerResult{
		Name: outputNameMp3,
		Path: filepath.Join(vc.PathOutput, outputNameMp3),
	}
	return res, nil
}

func (vc *VoiceChanger) RunMix(ctx context.Context, musicPath string, voicePath string) (VoiceChangerResult, error) {
	mixNameMp3 := fmt.Sprintf("Mix %s", filepath.Base(voicePath))

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return VoiceChangerResult{}, err
	}

	cmd := exec.CommandContext(ctx, ffmpeg,
		"-i", musicPath,
		"-i", voicePath,
		"-filter_complex", "[0:a]volume=0.5[a1];[1:a]volume=1[a2];[a1][a2]amix=inputs=2:duration=longest",
		"-b:a", "320k",
		"-y",
		filepath.Join(vc.PathOutput, mixNameMp3),
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err = cmd.Run()
	if err != nil && err.Error() == "signal: killed" {
		return VoiceChangerResult{}, context.Canceled
	} else if err != nil {
		return VoiceChangerResult{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
	}

	res := VoiceChangerResult{
		Name: mixNameMp3,
		Path: filepath.Join(vc.PathOutput, mixNameMp3),
	}
	return res, nil
}
