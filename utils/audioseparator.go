package utils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type AudioSeparator struct {
	Mode       string
	Model      string
	PathCLI    string
	PathModels string
	PathOutput string
}

func NewAudioSeparator() *AudioSeparator {
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &AudioSeparator{
			Mode:       "CUDA",
			Model:      "UVR-MDX-NET-Voc_FT",
			PathCLI:    filepath.Join(GetExecutablePath(), "audio-separator", "bin", "audio-separator"),
			PathModels: filepath.Join(GetExecutablePath(), "audio-separator", "models"),
			PathOutput: GetDownloadFolderPath(),
		}
	} else {
		return &AudioSeparator{
			Mode:       "CPU",
			Model:      "UVR-MDX-NET-Voc_FT",
			PathCLI:    filepath.Join(GetExecutablePath(), "audio-separator", "bin", "audio-separator"),
			PathModels: filepath.Join(GetExecutablePath(), "audio-separator", "models"),
			PathOutput: GetDownloadFolderPath(),
		}
	}
}

func (c *AudioSeparator) Run(ctx context.Context, file DownloadedFile) error {
	var cmd *exec.Cmd

	// "UVR-MDX-NET-Voc_FT" best for vocal
	//"UVR-MDX-NET-Inst_HQ_3" best for music

	switch c.Mode {
	case "CUDA":
		cmd = exec.CommandContext(ctx, c.PathCLI, file.Path,
			"--model_name", c.Model,
			"--model_file_dir", c.PathModels,
			"--output_dir", c.PathOutput,
			"--output_format=MP3",
			"--log_level=DEBUG",
			"--use_cuda",
		)
	default:
		cmd = exec.CommandContext(ctx, c.PathCLI, file.Path,
			"--model_name", c.Model,
			"--model_file_dir", c.PathModels,
			"--output_dir", c.PathOutput,
			"--output_format=MP3",
			"--log_level=DEBUG",
		)
	}

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	return cmd.Run()
}
