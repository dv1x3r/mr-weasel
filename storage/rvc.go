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
	UserID         int64          `db:"user_id"`
	ModelName      sql.NullString `db:"model_name"`
	DatasetFolder  sql.NullString `db:"dataset_folder"`
	ModelFile      sql.NullString `db:"model_file"`
	IndexFile      sql.NullString `db:"index_file"`
	AudioSourceID  sql.NullString `db:"audio_source_id"`
	AudioVoiceFile sql.NullString `db:"audio_voice_file"`
	AudioMusicFile sql.NullString `db:"audio_music_file"`
	SeparateUVR    sql.NullBool   `db:"separate_uvr"`
	Transpose      sql.NullInt64  `db:"transpose"`
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
			m.dataset_folder,
			m.model_file,
			m.index_file,
			e.audio_source_id,
			e.audio_voice_file,
			e.audio_music_file,
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

func (s *RvcStorage) SetExperimentAudioSourceInDB(ctx context.Context, userID int64, experimentID int64, audioSourceID string) error {
	stmt := `update rvc_experiment set audio_source_id = ? where user_id = ? and id = ?;`
	_, err := s.db.ExecContext(ctx, stmt, audioSourceID, userID, experimentID)
	return err
}

func (s *RvcStorage) SetExperimentAudioPathInDB(ctx context.Context, userID int64, experimentID int64, audioVoicePath sql.NullString, audioMusicPath sql.NullString) error {
	stmt := `update rvc_experiment set audio_voice_path = ?, audio_music_path = ? where user_id = ? and id = ?;`
	_, err := s.db.ExecContext(ctx, stmt, audioVoicePath, audioMusicPath, userID, experimentID)
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
			m.id
			,m.name
			,iif(a.id is null, 1, 0) as is_owner
			,coalesce(ac.shares, 0) as shares
			,count(*) over () as countrows
		from rvc_model m
		left join rvc_access a on a.model_id = m.id
		left join (
			select
				model_id,
				count(*) as shares
			from rvc_access
			group by model_id
		) ac on ac.model_id = m.id
		where m.user_id = ? or a.user_id = ?
		order by m.name, m.id
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &model, stmt, userID, userID, offset)
	return model, err
}