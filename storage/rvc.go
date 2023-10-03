package storage

import (
	// "context"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type RvcStorage struct {
	db *sqlx.DB
}

func NewRvcStorage(db *sqlx.DB) *RvcStorage {
	return &RvcStorage{db: db}
}

type RvcExperimentDetails struct {
	UserID    int64          `db:"user_id"`
	ModelName sql.NullString `db:"model_name"`
	AudioPath sql.NullString `db:"audio_path"`
	EnableUVR sql.NullInt64  `db:"enable_uvr"`
	Transpose sql.NullInt64  `db:"transpose"`
}

func (s *RvcStorage) NewExperiment(ctx context.Context, userID int64) (int64, error) {
	stmt := "insert into rvc_experiment (user_id, enable_uvr, transpose) values (?,0,0);"
	res, err := s.db.ExecContext(ctx, stmt, userID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *RvcStorage) GetExperimentDetails(ctx context.Context, userID int64, experimentID int64) (RvcExperimentDetails, error) {
	var experiment RvcExperimentDetails
	stmt := `
		select e.user_id, m.name as model_name, e.audio_path, e.enable_uvr, e.transpose
		from rvc_experiment e
		left join rvc_model m on m.id = e.model_id
		where e.user_id = ? and e.id = ?;
	`
	err := s.db.GetContext(ctx, &experiment, stmt, userID, experimentID)
	return experiment, err
}
