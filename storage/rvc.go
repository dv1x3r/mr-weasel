package storage

import (
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
	UserID      int64          `db:"user_id"`
	ModelName   sql.NullString `db:"model_name"`
	Audio       sql.NullString `db:"audio"`
	SeparateUVR sql.NullBool   `db:"separate_uvr"`
	Transpose   sql.NullInt64  `db:"transpose"`
}

func (s *RvcStorage) InsertNewExperimentIntoDB(ctx context.Context, userID int64) (int64, error) {
	stmt := "insert into rvc_experiment (user_id) values (?);"
	res, err := s.db.ExecContext(ctx, stmt, userID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *RvcStorage) GetExperimentDetailsFromDB(ctx context.Context, userID int64, experimentID int64) (RvcExperimentDetails, error) {
	var experiment RvcExperimentDetails
	stmt := `
		select
			e.user_id,
			m.name as model_name,
			e.audio,
			e.separate_uvr,
			e.transpose
		from rvc_experiment e
		left join rvc_model m on m.id = e.model_id
		where e.user_id = ? and e.id = ?;
	`
	err := s.db.GetContext(ctx, &experiment, stmt, userID, experimentID)
	return experiment, err
}

func (s *RvcStorage) SetExperimentModelInDB(ctx context.Context, userID int64, experimentID int64, modelID int64) error {
	stmt := `
		select m.id
		from rvc_model m
		left join rvc_access a on a.model_id = m.id
		where m.user_id = ? or a.user_id = ?;
	`
	var check int64
	err := s.db.GetContext(ctx, &check, stmt, userID, userID)
	if err != nil {
		return err
	}

	stmt = `update rvc_experiment set model_id = ? where user_id = ? and id = ?;`
	_, err = s.db.ExecContext(ctx, stmt, modelID, userID, experimentID)
	return err
}

func (s *RvcStorage) SetExperimentAudioInDB(ctx context.Context, userID int64, experimentID int64, audio string) error {
	stmt := `update rvc_experiment set audio = ? where user_id = ? and id = ?;`
	_, err := s.db.ExecContext(ctx, stmt, audio, userID, experimentID)
	return err
}

func (s *RvcStorage) SetExperimentSeparateUVRInDB(ctx context.Context, userID int64, experimentID int64, separateUVR bool) error {
	stmt := `update rvc_experiment set separate_uvr = ? where user_id = ? and id = ?;`
	_, err := s.db.ExecContext(ctx, stmt, separateUVR, userID, experimentID)
	return err
}

func (s *RvcStorage) SetExperimentTransposeInDB(ctx context.Context, userID int64, experimentID int64, transpose int64) error {
	stmt := `update rvc_experiment set transpose = ? where user_id = ? and id = ?;`
	_, err := s.db.ExecContext(ctx, stmt, transpose, userID, experimentID)
	return err
}

type RvcModelDetails struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	IsOwner   bool   `db:"is_owner"`
	Shares    int64  `db:"shares"`
	CountRows int64  `db:"countrows"`
}

func (s *RvcStorage) GetModelFromDB(ctx context.Context, userID int64, offset int64) (RvcModelDetails, error) {
	var model RvcModelDetails
	stmt := `
		select
			t.id,
			t.name,
			t.is_owner,
			coalesce(s.shares, 0) as shares,
			count(*) over() as countrows
		from (
			select
				m.id,
				m.user_id,
				m.name,
				1 as is_owner
			from rvc_model m
			union
			select
				m.id,
				a.user_id,
				m.name,
				0 as is_owner
			from rvc_model m
			join rvc_access a on a.model_id = m.id
			where m.user_id <> a.user_id -- safety check
		) t
		left join (
			select
				model_id,
				count(*) as shares
			from rvc_access
			group by model_id
		) s on s.model_id = t.id
		where t.user_id = ?
		order by t.name, t.id
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &model, stmt, userID, offset)
	return model, err
}

func (s *RvcStorage) InsertNewModelIntoDB(ctx context.Context, userID int64, name string) (int64, error) {
	stmt := `insert into rvc_model (user_id, name) values (?,?);`
	res, err := s.db.ExecContext(ctx, stmt, userID, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *RvcStorage) DeleteModelFromDB(ctx context.Context, userID int64, modelID int64) (int64, error) {
	stmt := `delete from rvc_model where id = ? and user_id = ?;`
	res, err := s.db.ExecContext(ctx, stmt, modelID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *RvcStorage) InsertNewAccessIntoDB(ctx context.Context, userID int64, modelID int64, accessUserID int64) (int64, error) {
	stmt := `select id from rvc_model where user_id = ? and id = ?;`
	var check int64
	err := s.db.GetContext(ctx, &check, stmt, userID, modelID)
	if err != nil {
		return 0, err
	}

	stmt = `insert into rvc_access (user_id, model_id) values (?,?);`
	res, err := s.db.ExecContext(ctx, stmt, accessUserID, modelID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *RvcStorage) DeleteAccessFromDB(ctx context.Context, userID int64, modelID int64) (int64, error) {
	stmt := `
		delete from rvc_access where model_id = ?
			and model_id in (select id from rvc_model where user_id = ?);
	`
	res, err := s.db.ExecContext(ctx, stmt, modelID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
