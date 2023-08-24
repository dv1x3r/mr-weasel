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

type HolidayDaysByYear struct {
	Year int64 `db:"year"`
	Days int64 `db:"days"`
}

func (h *HolidayBase) GetStartTimestamp() string {
	return time.Unix(h.Start, 0).UTC().Format("Monday, 02 January 2006")
}

func (h *HolidayBase) GetEndTimestamp() string {
	return time.Unix(h.End, 0).UTC().Format("Monday, 02 January 2006")
}

func (s *HolidayStorage) SelectHolidayDaysByYearFromDB(ctx context.Context, userID int64) ([]HolidayDaysByYear, error) {
	var holidays []HolidayDaysByYear
	stmt := `
		select year(start) as year, sum(days) as days
		from holiday
		where user_id = ?
		order by 1;
	`
	err := s.db.SelectContext(ctx, &holidays, stmt, userID)
	return holidays, err
}

func (s *CarStorage) GetHolidayFromDB(ctx context.Context, userID int64, offset int64) (HolidayBase, error) {
	var holiday HolidayBase
	stmt := `
		select id, user_id, start, end, days
		from holidays
		where user_id = ?
		order by start desc
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &holiday, stmt, userID, offset)
	return holiday, err
}
