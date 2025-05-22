package db

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"modernc.org/sqlite"
	sqliteLib "modernc.org/sqlite/lib"
)

type SQLiteStore struct {
	db *sqlx.DB
}

func NewSQLiteStore(filePath string) (*SQLiteStore, error) {
	const args = "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)"
	db, err := sqlx.Connect("sqlite", filePath+args)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s SQLiteStore) DB() *sql.DB {
	return s.db.DB
}

func (s SQLiteStore) DBX() *sqlx.DB {
	return s.db
}

func (SQLiteStore) IsErrConstraintUnique(err error) bool {
	var sqliteErr *sqlite.Error
	ok := errors.As(err, &sqliteErr)
	return ok && sqliteErr.Code() == sqliteLib.SQLITE_CONSTRAINT_UNIQUE
}

func (SQLiteStore) IsErrConstraintTrigger(err error) bool {
	var sqliteErr *sqlite.Error
	ok := errors.As(err, &sqliteErr)
	return ok && sqliteErr.Code() == sqliteLib.SQLITE_CONSTRAINT_TRIGGER
}

func (SQLiteStore) IsErrConstraintForeignKey(err error) bool {
	var sqliteErr *sqlite.Error
	ok := errors.As(err, &sqliteErr)
	return ok && sqliteErr.Code() == sqliteLib.SQLITE_CONSTRAINT_FOREIGNKEY
}
