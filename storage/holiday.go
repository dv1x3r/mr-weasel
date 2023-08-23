package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type HolidayStorage struct {
	db *sqlx.DB
}

func NewHolidayStorage(db *sqlx.DB) *HolidayStorage {
	return &HolidayStorage{db: db}
}

type HolidayBase struct {
	ID     int64 `db:"id"`
	UserID int64 `db:"user_id"`
	Start  int64 `db:"start"`
	End    int64 `db:"end"`
	Days   int64 `db:"days"`
}

func (h *HolidayBase) GetStartTimestamp() string {
	return time.Unix(h.Start, 0).UTC().Format("Monday, 02 January 2006")
}

func (h *HolidayBase) GetEndTimestamp() string {
	return time.Unix(h.End, 0).UTC().Format("Monday, 02 January 2006")
}

func (s *CarStorage) SelectHolidaysFromDB(ctx context.Context, userID int64) ([]HolidayBase, error) {
	var holidays []HolidayBase
	stmt := `
		select id, user_id, start, end, days
		from holiday
		where user_id = ?
		order by start;
	`
	err := s.db.SelectContext(ctx, &holidays, stmt, userID)
	return holidays, err
}
