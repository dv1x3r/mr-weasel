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

	"github.com/jmoiron/sqlx"
)

const downloadPath = "downloads/"

type Blob struct {
	db         *sqlx.DB
	httpClient *http.Client
}

func NewBlob(db *sqlx.DB) *Blob {
	os.MkdirAll(downloadPath, os.ModePerm)
	return &Blob{db: db, httpClient: &http.Client{}}
}

type BlobBase struct {
	ID          int64  `db:"id"`
	UserID      int64  `db:"user_id"`
	FileID      string `db:"file_id"`
	Extension   string `db:"extension"`
	Description string `db:"description"`
}

func (b *BlobBase) GetAbsolutePath() string {
	return filepath.Join(downloadPath, fmt.Sprintf("%d.%s", b.ID, b.Extension))
}

type BlobPayload struct {
	FileID      string
	Description string
	URL         string
}

func (b *Blob) GetBlobFromDB(ctx context.Context, userID int64, blobID int64) (BlobBase, error) {
	var blob BlobBase
	stmt := `select id, user_id, file_id, extension, description from blob where user_id = ? and id = ?;`
	err := b.db.GetContext(ctx, &blob, stmt, userID, blobID)
	return blob, err
}

func (b *Blob) InsertBlobIntoDB(ctx context.Context, blob BlobBase) (int64, error) {
	stmt := "insert into blob (user_id, file_id, extension, description) values (?,?,?,?);"
	res, err := b.db.ExecContext(ctx, stmt, blob.UserID, blob.FileID, blob.Extension, blob.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (b *Blob) DownloadBlob(ctx context.Context, userID int64, blobPayload *BlobPayload) (BlobBase, error) {
	URL, err := url.Parse(blobPayload.URL)
	if err != nil {
		return BlobBase{}, err
	}
	if URL.Hostname() != "api.telegram.org" {
		return BlobBase{}, errors.New("not api.telegram.org blob")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", URL.String(), nil)
	if err != nil {
		return BlobBase{}, err
	}
	res, err := b.httpClient.Do(req)
	if err != nil {
		return BlobBase{}, err
	}
	defer res.Body.Close()

	blobID, err := b.InsertBlobIntoDB(ctx, BlobBase{
		UserID:      userID,
		FileID:      blobPayload.FileID,
		Extension:   filepath.Ext(URL.String()),
		Description: blobPayload.Description,
	})
	if err != nil {
		return BlobBase{}, err
	}

	blob, err := b.GetBlobFromDB(ctx, userID, blobID)
	if err != nil {
		return blob, err
	}

	blobFile, err := os.Create(blob.GetAbsolutePath())
	if err != nil {
		return blob, err
	}
	defer blobFile.Close()

	_, err = io.Copy(blobFile, res.Body)
	return blob, err
}

func (b *Blob) DownloadYouTube(ctx context.Context, userID int64, link string) (BlobBase, error) {
	return BlobBase{}, errors.New("not implemented")
}
