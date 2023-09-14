package utils

import (
	"context"
	// "io/fs"
	"net/http"
	// "net/url"
	// "os"
	// "path/filepath"

	// "mr-weasel/commands"

	"github.com/jmoiron/sqlx"
)

type Blob struct {
	db         *sqlx.DB
	httpClient *http.Client
}

func NewBlob(db *sqlx.DB) *Blob {
	return &Blob{db: db, httpClient: &http.Client{}}
}

type BlobBase struct {
	ID     int64  `db:"id"`
	UserID int64  `db:"user_id"`
	FileID string `db:"file_id"`
}

func (b *Blob) GetBlobFromDB(ctx context.Context, userID int64, blobID int64) (BlobBase, error) {
	var blob BlobBase
	stmt := `select id, user_id, file_id from blob where user_id = ? and id = ?;`
	err := b.db.GetContext(ctx, &blob, stmt, userID, blobID)
	return blob, err
}

func (b *Blob) InsertBlobIntoDB(ctx context.Context, blob BlobBase) (int64, error) {
	stmt := "insert into blob (user_id, file_id) values (?,?);"
	res, err := b.db.ExecContext(ctx, stmt, blob.UserID, blob.FileID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (b *Blob) DownloadBlob(ctx context.Context, userID int64, blobID int64) {

}

func (b *Blob) UploadBlob(ctx context.Context, userID int64, blobID int64) {

}

// func (b *Blob) DownloadUrl(ctx context.Context, pl commands.Payload) (BlobBase, error) {
// parsedURL, err := url.Parse(fullURL)
// if err != nil {
// 	return BlobBase{}, err
// }

// if parserURL.Hostname() == "api.telegram.org" {
// 	resp, err := b.httpClient.Get(fullURL)
// 	if err == nil {
// 		return BlobBase{}, err
// 	}
// 	defer resp.Body.Close()
// 	// b.InsertBlobIntoDB(ctx, BlobBase{UserID: userID, FileID: })

// } else {

// }

// blobPath := filepath.Join(b.blobPath, blobName)

// file, err := os.Create(blobPath)
// // 	if err != nil {
// // 		pl.ResultChan <- Result{Text: "Unable to save the file on the server, please try again :c", State: c.receiveURL, Error: err}
// 		return
// 	}
// 	defer file.Close()
// 	size, err := io.Copy(file, resp.Body)
// 	pl.ResultChan <- Result{Text: "File has been received! size " + fmt.Sprintf("%d", size)}

// return BlobBase{}, nil
// }
