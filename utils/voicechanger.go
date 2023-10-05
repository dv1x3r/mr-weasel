package utils

import (
	"context"
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
	ResultName string
	ResultPath string
}

func NewVoiceChanger() *VoiceChanger {
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &VoiceChanger{
			Mode:         "CUDA",
			PathPython:   "/mnt/d/rvc-project/.venv/Scripts/python.exe",
			PathInferCLI: "/mnt/d/rvc-project/infer-cli.py",
			PathTrainCLI: "/mnt/d/rvc-project/train-cli.py",
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

func (vc *VoiceChanger) RunTrain(ctx context.Context, experiment storage.RvcExperimentDetails) error {

	return nil
}

func (vc *VoiceChanger) RunInfer(ctx context.Context, experiment storage.RvcExperimentDetails) (VoiceChangerResult, error) {

	return VoiceChangerResult{}, nil
}

func (vc *VoiceChanger) RunMix(ctx context.Context, experiment storage.RvcExperimentDetails) (VoiceChangerResult, error) {

	return VoiceChangerResult{}, nil
}
