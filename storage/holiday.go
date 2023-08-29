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

type HolidayDetails struct {
	HolidayBase
	CountRows int64 `db:"countrows"`
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
		select cast(strftime('%Y', start, 'unixepoch') as integer) as year, sum(days) as days
		from holiday
		where user_id = ?
		group by 1 order by 1;
	`
	err := s.db.SelectContext(ctx, &holidays, stmt, userID)
	return holidays, err
}

func (s *HolidayStorage) GetHolidayFromDB(ctx context.Context, userID int64, offset int64) (HolidayDetails, error) {
	var holiday HolidayDetails
	stmt := `
		select id, user_id, start, end, days
			,count(*) over (partition by user_id) as countrows
		from holiday
		where user_id = ?
		order by start desc
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &holiday, stmt, userID, offset)
	return holiday, err
}

func (s *HolidayStorage) InsertHolidayIntoDB(ctx context.Context, holiday HolidayBase) (int64, error) {
	stmt := "insert into holiday (user_id, start, end, days) values (?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, holiday.UserID, holiday.Start, holiday.End, holiday.Days)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *HolidayStorage) DeleteHolidayFromDB(ctx context.Context, userID int64, holidayID int64) (int64, error) {
	stmt := `delete from holiday where user_id = ? and id = ?;`
	res, err := s.db.ExecContext(ctx, stmt, userID, holidayID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
