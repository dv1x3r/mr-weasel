package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AudioSeparator struct {
	Mode       string
	Model      string
	PathCLI    string
	PathModels string
	PathOutput string
}

type AudioSeparatorResult struct {
	MusicName string
	MusicPath string
	VoiceName string
	VoicePath string
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

func (c *AudioSeparator) Run(ctx context.Context, file DownloadedFile) (AudioSeparatorResult, error) {
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

	err := cmd.Run()
	if err != nil && err.Error() == "signal: killed" {
		return AudioSeparatorResult{}, context.Canceled
	} else if err != nil {
		return AudioSeparatorResult{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
	}

	baseName := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Name))

	res := AudioSeparatorResult{
		MusicName: fmt.Sprintf("Instrumental_%s", file.Name),
		MusicPath: filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Instrumental)_%s.mp3", baseName, c.Model)),
		VoiceName: fmt.Sprintf("Vocals_%s", file.Name),
		VoicePath: filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Vocals)_%s.mp3", baseName, c.Model)),
	}

	return res, nil
}

func (c *AudioSeparator) Exists(file DownloadedFile) bool {
	baseName := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Name))
	musicPath := filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Instrumental)_%s.mp3", baseName, c.Model))
	voicePath := filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Vocals)_%s.mp3", baseName, c.Model))
	_, err1 := os.Stat(musicPath)
	_, err2 := os.Stat(voicePath)
	return errors.Join(err1, err2) == nil
}
