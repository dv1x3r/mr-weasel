package utils

import (
	"os/exec"
	"path/filepath"
)

type VoiceChanger struct {
	separator    *AudioSeparator
	Mode         string
	PathPython   string
	PathInferCLI string
	PathTrainCLI string
	PathDatasets string
	PathWeights  string
	PathLogs     string
}

func NewVoiceChanger() *VoiceChanger {
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &VoiceChanger{
			separator:    &AudioSeparator{},
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
			separator:    &AudioSeparator{},
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
