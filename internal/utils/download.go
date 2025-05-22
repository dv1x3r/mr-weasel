package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetExecutablePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func GetDownloadFolderPath() string {
	return filepath.Join(GetExecutablePath(), "temp")
}

type DownloadedFile struct {
	ID   string
	Name string
	Path string
}

func GetDownloadedFile(fileID string) (DownloadedFile, error) {
	folderPath := GetDownloadFolderPath()

	dir, err := os.Open(folderPath)
	if err != nil {
		return DownloadedFile{}, err
	}
	defer dir.Close()

	files, err := dir.ReadDir(-1)
	if err != nil {
		return DownloadedFile{}, err
	}

	for _, file := range files {
		split := strings.SplitN(file.Name(), ".", 2)
		if len(split) == 2 && split[0] == fileID {
			return DownloadedFile{
				ID:   fileID,
				Name: split[1],
				Path: filepath.Join(folderPath, file.Name()),
			}, nil
		}
	}

	return DownloadedFile{}, errors.New("downloaded file not found")
}

func Download(ctx context.Context, arg1 string, arg2 string) (DownloadedFile, error) {
	var rawURL, fileName string

	if arg1 != "" {
		rawURL = arg1
		fileName = arg2
	} else {
		rawURL = arg2
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return DownloadedFile{}, err
	}

	b := make([]byte, 16)
	rand.Read(b)
	fileID := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	downloadFolderPath := GetDownloadFolderPath()
	os.MkdirAll(downloadFolderPath, os.ModePerm)

	if parsedURL.Hostname() == "api.telegram.org" {
		req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
		if err != nil {
			return DownloadedFile{}, err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return DownloadedFile{}, err
		}
		defer res.Body.Close()

		file, err := os.Create(filepath.Join(downloadFolderPath, fmt.Sprintf("%s.%s", fileID, fileName)))
		if err != nil {
			return DownloadedFile{}, err
		}
		defer file.Close()

		filePath := file.Name()

		if _, err = io.Copy(file, res.Body); err != nil {
			return DownloadedFile{}, err
		}

		if strings.HasSuffix(fileName, ".oga") {
			fileNameNew := fmt.Sprintf("%s.mp3", strings.TrimSuffix(fileName, ".oga"))
			filePathNew := filepath.Join(filepath.Dir(filePath), fmt.Sprintf("%s.%s", fileID, fileNameNew))

			ffmpeg, err := exec.LookPath("ffmpeg")
			if err != nil {
				return DownloadedFile{}, err
			}

			cmd := exec.Command(ffmpeg, "-i", filePath, "-b:a", "320k", "-y", filePathNew)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

			err = cmd.Run()
			if err != nil {
				return DownloadedFile{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
			}

			os.Remove(filePath)
			filePath = filePathNew
			fileName = fileNameNew
		}

		df := DownloadedFile{
			ID:   fileID,
			Name: fileName,
			Path: filePath,
		}

		return df, err

	} else {
		dlp, err := exec.LookPath("yt-dlp")
		if err != nil {
			return DownloadedFile{}, err
		}

		cmd := exec.CommandContext(ctx, dlp, rawURL,
			"-x",
			"--audio-format=mp3",
			"--audio-quality=0",
			"--max-filesize=50M",
			"--playlist-items=1",
			"--paths", downloadFolderPath,
			"--output", fileID+".%(title)s.mp3",
			"--print=after_move:title",
		)
		cmd.Stdout, cmd.Stderr = &bytes.Buffer{}, &bytes.Buffer{}

		err = cmd.Run()
		if err != nil && err.Error() == "signal: killed" {
			return DownloadedFile{}, context.Canceled
		} else if err != nil {
			return DownloadedFile{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
		}

		title := strings.TrimSpace(fmt.Sprint(cmd.Stdout))
		if title == "" {
			return DownloadedFile{}, errors.New("yt-dlp outputed an empty video title")
		}

		df := DownloadedFile{
			ID:   fileID,
			Name: fmt.Sprintf("%s.mp3", title),
			Path: filepath.Join(downloadFolderPath, fmt.Sprintf("%s.%s.mp3", fileID, title)),
		}

		return df, nil
	}
}
