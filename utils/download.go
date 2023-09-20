package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

func GetDownloadOriginalName(fileBase string) string {
	split := strings.SplitN(fileBase, ".", 2)
	if len(split) == 2 {
		return split[1]
	} else {
		return fileBase
	}
}

func Download(ctx context.Context, rawURL string, name string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsedURL.Hostname() != "api.telegram.org" {
		return "", errors.New("not api.telegram.org blob")
	} else {
		req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
		if err != nil {
			return "", err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		folderPath := GetDownloadFolderPath()

		os.MkdirAll(GetDownloadFolderPath(), os.ModePerm)

		file, err := os.CreateTemp(folderPath, fmt.Sprintf("*.%s", name))
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = io.Copy(file, res.Body)
		return filepath.Join(folderPath, file.Name()), err
	}
}
