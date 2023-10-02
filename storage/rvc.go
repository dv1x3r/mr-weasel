package storage

import (
	// "context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type RvcStorage struct {
	db *sqlx.DB
}

func NewRvcStorage(db *sqlx.DB) *RvcStorage {
	return &RvcStorage{db: db}
}

type RvcExperimentBase struct {
	ID        int64          `db:"id"`
	UserID    int64          `db:"user_id"`
	ModelID   sql.NullInt64  `db:"model_id"`
	AudioFile sql.NullString `db:"audio_file"`
	EnableUVR sql.NullInt64  `db:"enable_uvr"`
	Transpose sql.NullInt64  `db:"transpose"`
}
