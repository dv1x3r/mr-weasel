package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

func (vc *VoiceChanger) RunTrain(ctx context.Context, modelID int64) error {
	modelName := fmt.Sprint(modelID)

	var cmd *exec.Cmd

	switch vc.Mode {
	case "CUDA":
		cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathTrainCLI,
			"--name", modelName,
			"--dataset", filepath.Join("assets", "datasets", modelName),
			"--version", "v2",
			"--sample_rate", "40k",
			"--method", "rmvpe_gpu",
			"--gpu_rmvpe", "0-0",
			"--gpu", "0",
			"--batch_size", "8",
			"--total_epoch", "200",
			"--save_epoch", "20",
			"--save_latest", "1",
			"--cache_gpu", "0",
			"--save_every_weights", "0",
		)
	default:
		cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathTrainCLI,
			"--name", modelName,
			"--dataset", filepath.Join("assets", "datasets", modelName),
			"--version", "v2",
			"--sample_rate", "40k",
			"--method", "rmvpe",
			// "--gpu_rmvpe", "0-0",
			// "--gpu", "0",
			"--batch-size", "1",
			"--total_epoch", "10",
			"--save_epoch", "5",
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
		filepath.Join(vc.PathLogs, modelName, "added_IVF136_Flat_nprobe_1_3_v2.index"),
		filepath.Join(vc.PathWeights, fmt.Sprintf("%s.index", modelName)),
	)

	// TODO: clear training data after some tests
	// os.RemoveAll(filepath.Join(vc.PathLogs, modelName))
	// os.RemoveAll(filepath.Join(vc.PathDatasets, modelName))

	return nil
}

func (vc *VoiceChanger) RunInfer(ctx context.Context, experiment storage.RvcExperimentDetails, uvrFiles AudioSeparatorResult) (VoiceChangerResult, error) {
	// modelName := fmt.Sprint(experiment.ModelID)

	// var cmd *exec.Cmd

	// switch vc.Mode {
	// case "CUDA":
	// 	cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathTrainCLI,
	// 		"--name", modelName,
	// 		"--dataset", filepath.Join("assets", "datasets", modelName),
	// 		"--version", "v2",
	// 		"--sample_rate", "40k",
	// 		"--method", "rmvpe_gpu",
	// 		"--gpu_rmvpe", "0-0",
	// 		"--gpu", "0",
	// 		"--batch_size", "8",
	// 		"--total_epoch", "200",
	// 		"--save_epoch", "20",
	// 		"--save_latest", "1",
	// 		"--cache_gpu", "0",
	// 		"--save_every_weights", "0",
	// 	)
	// default:
	// 	cmd = exec.CommandContext(ctx, vc.PathPython, vc.PathTrainCLI,
	// 		"--name", modelName,
	// 		"--dataset", filepath.Join("assets", "datasets", modelName),
	// 		"--version", "v2",
	// 		"--sample_rate", "40k",
	// 		"--method", "rmvpe",
	// 		// "--gpu_rmvpe", "0-0",
	// 		// "--gpu", "0",
	// 		"--batch-size", "1",
	// 		"--total_epoch", "10",
	// 		"--save_epoch", "5",
	// 		"--save_latest", "1",
	// 		"--cache_gpu", "0",
	// 		"--save_every_weights", "0",
	// 	)
	// }

	// cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	// err := cmd.Run()
	// if err != nil && err.Error() == "signal: killed" {
	// 	return VoiceChangerResult{}, context.Canceled
	// } else if err != nil {
	// 	return VoiceChangerResult{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
	// }

	return VoiceChangerResult{}, nil
}

// TODO:
// func (vc *VoiceChanger) RunMix(ctx context.Context, experiment storage.RvcExperimentDetails) (VoiceChangerResult, error) {
// 	return VoiceChangerResult{}, nil
// }
