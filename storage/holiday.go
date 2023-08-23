package storage

import "github.com/jmoiron/sqlx"

type HolidayStorage struct {
	db *sqlx.DB
}

func NewHolidayStorage(db *sqlx.DB) *HolidayStorage {
	return &HolidayStorage{db: db}
}
