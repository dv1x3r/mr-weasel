package utils

import (
	"bytes"
	"context"
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
	UniqueID string
	Name     string
	Path     string
}

func GetDownloadedFile(uniqueID string) (DownloadedFile, error) {
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
		fileName := file.Name()
		split := strings.Split(fileName, ".")
		if len(split) > 1 && split[len(split)-2] == uniqueID {
			return DownloadedFile{
				UniqueID: split[len(split)-2],
				Name:     strings.Join(split[0:len(split)-2], "."),
				Path:     filepath.Join(folderPath, fileName),
			}, nil
		}
	}

	return DownloadedFile{}, errors.New("downloaded file not found")
}

func Download(ctx context.Context, rawURL string, fileName string) (DownloadedFile, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return DownloadedFile{}, err
	}

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

		file, err := os.CreateTemp(downloadFolderPath, fmt.Sprintf("%s.*.%s", filepath.Base(fileName), filepath.Ext(fileName)))
		if err != nil {
			return DownloadedFile{}, err
		}
		defer file.Close()

		_, err = io.Copy(file, res.Body)

		nameSplit := strings.Split(filepath.Base(file.Name()), ".")
		df := DownloadedFile{
			UniqueID: nameSplit[len(nameSplit)-1],
			Name:     fileName,
			Path:     filepath.Join(downloadFolderPath, file.Name()),
		}

		return df, err

	} else {

		uniqueID := fmt.Sprint(ctx.Value("contextID"))

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
			"--output=%(title)s."+uniqueID+".mp3",
			"--print=after_move:title",
		)
		cmd.Stdout, cmd.Stderr = &bytes.Buffer{}, &bytes.Buffer{}

		err = cmd.Run()
		if err != nil {
			return DownloadedFile{}, fmt.Errorf("%w: %s", err, cmd.Stderr)
		}

		title := strings.TrimSpace(fmt.Sprint(cmd.Stdout))
		if title == "" {
			return DownloadedFile{}, errors.New("yt-dlp outputed an empty video title")
		}

		df := DownloadedFile{
			UniqueID: uniqueID,
			Name:     fmt.Sprintf("%s.mp3", title),
			Path:     filepath.Join(downloadFolderPath, fmt.Sprintf("%s.%s.mp3", title, uniqueID)),
		}

		return df, nil
	}
}
