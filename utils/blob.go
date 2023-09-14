package utils

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type Blob struct {
	db         *sqlx.DB
	httpClient *http.Client
}

func NewBlob(db *sqlx.DB) *Blob {
	return &Blob{db: db, httpClient: &http.Client{}}
}

func (b *Blob) DownloadTelegramAudioIntoBlob(url string) (int64, error) {
	// 1. user uploads the song
	// check if url starts with https://api.telegram.org/
	// if yes, then just download the file
	// if not, then try to use yt-dlp
	// insert into blob (id, user_id, file_id, is_deleted, uploaded_at);

	// parsedURL, err := url.Parse(pl.FileURL)
	// if err != nil {
	// 	pl.ResultChan <- Result{Text: "Invalid link format, can you try another one?", State: c.receiveURL}
	// 	return
	// }

	// if parsedURL.Hostname() == "api.telegram.org" {
	// 	resp, err := c.httpClient.Get(pl.FileURL)
	// 	if err != nil {
	// 		pl.ResultChan <- Result{Text: "Unable to download the file, please try again :c", State: c.receiveURL, Error: err}
	// 		return
	// 	}
	// 	defer resp.Body.Close()

	// 	blobID := 0 // temp

	// 	extension := filepath.Ext(parsedURL.Path)
	// 	blobName := fmt.Sprintf("%d%s", blobID, extension)
	// 	blobPath := filepath.Join("blob", blobName)

	// 	_ = os.Mkdir("blob", fs.ModePerm)
	// 	file, err := os.Create(blobPath)
	// 	if err != nil {
	// 		pl.ResultChan <- Result{Text: "Unable to save the file on the server, please try again :c", State: c.receiveURL, Error: err}
	// 		return
	// 	}
	// 	defer file.Close()
	// 	size, err := io.Copy(file, resp.Body)
	// 	pl.ResultChan <- Result{Text: "File has been received! size " + fmt.Sprintf("%d", size)}

	// } else {
	// 	// TODO: yt-dlp
	// 	pl.ResultChan <- Result{Text: "sorry, links are not supported atm, please send me audio file", State: c.receiveURL}
	// 	return
	// }

	return 0, nil
}

func (b *Blob) DownloadYouTubeAudioIntoBlob(url string) (int64, error) {
	return 0, nil
}
