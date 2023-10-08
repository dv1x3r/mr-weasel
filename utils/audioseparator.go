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

// "UVR-MDX-NET-Voc_FT" best for vocal
//"UVR-MDX-NET-Inst_HQ_3" best for music

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
			PathOutput: filepath.Join(GetExecutablePath(), "audio-separator", "output"),
		}
	} else {
		return &AudioSeparator{
			Mode:       "CPU",
			Model:      "UVR-MDX-NET-Voc_FT",
			PathCLI:    filepath.Join(GetExecutablePath(), "audio-separator", "bin", "audio-separator"),
			PathModels: filepath.Join(GetExecutablePath(), "audio-separator", "models"),
			PathOutput: filepath.Join(GetExecutablePath(), "audio-separator", "output"),
		}
	}
}

func (c *AudioSeparator) Run(ctx context.Context, file DownloadedFile) (AudioSeparatorResult, error) {
	baseName := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Name))
	res := AudioSeparatorResult{
		MusicName: fmt.Sprintf("Instrumental_%s", file.Name),
		MusicPath: filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Instrumental)_%s.mp3", baseName, c.Model)),
		VoiceName: fmt.Sprintf("Vocals_%s", file.Name),
		VoicePath: filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Vocals C)_%s.mp3", baseName, c.Model)),
	}

	_, err1 := os.Stat(res.MusicPath)
	_, err2 := os.Stat(res.VoicePath)

	exists := errors.Join(err1, err2) == nil
	if exists {
		return res, nil
	}

	var cmd *exec.Cmd

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

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return AudioSeparatorResult{}, err
	}

	cmd = exec.CommandContext(ctx, ffmpeg,
		"-i", filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Vocals)_%s.mp3", baseName, c.Model)),
		"-filter_complex", "compand=attacks=0:points=-80/-900|-45/-15|-27/-9|0/-7|20/-7:gain=5",
		"-b:a", "320k",
		"-y",
		filepath.Join(c.PathOutput, fmt.Sprintf("%s_(Vocals C)_%s.mp3", baseName, c.Model)),
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err = cmd.Run()
	if err != nil && err.Error() == "signal: killed" {
		return AudioSeparatorResult{}, context.Canceled
	} else if err != nil {
		return AudioSeparatorResult{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
	}

	return res, nil
}
