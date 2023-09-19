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
)

const DownloadDir = "downloads"

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

		file, err := os.CreateTemp(DownloadDir, fmt.Sprintf("*.%s", name))
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = io.Copy(file, res.Body)
		return filepath.Join("downloads", file.Name()), err
	}
}
