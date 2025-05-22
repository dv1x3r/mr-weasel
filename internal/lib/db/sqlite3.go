package db

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

type SQLite3Store struct {
	db *sqlx.DB
}

func NewSQLite3Store(filePath string) (*SQLite3Store, error) {
	const args = "?_journal=WAL&_fk=1&_busy_timeout=10000"
	db, err := sqlx.Connect("sqlite3", filePath+args)
	if err != nil {
		return nil, err
	}
	return &SQLite3Store{db: db}, nil
}

func (s *SQLite3Store) DB() *sql.DB {
	return s.db.DB
}

func (s *SQLite3Store) DBX() *sqlx.DB {
	return s.db
}

func (SQLite3Store) IsErrConstraintUnique(err error) bool {
	var sqliteErr sqlite3.Error
	ok := errors.As(err, &sqliteErr)
	return ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}

func (SQLite3Store) IsErrConstraintTrigger(err error) bool {
	var sqliteErr sqlite3.Error
	ok := errors.As(err, &sqliteErr)
	return ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintTrigger
}

func (SQLite3Store) IsErrConstraintForeignKey(err error) bool {
	var sqliteErr sqlite3.Error
	ok := errors.As(err, &sqliteErr)
	return ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey
}
