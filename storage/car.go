package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type CarStorage struct {
	db *sqlx.DB
}

func NewCarStorage(db *sqlx.DB) *CarStorage {
	return &CarStorage{db: db}
}

type Car struct {
	ID     int64   `db:"id"`
	UserID int64   `db:"user_id"`
	Name   string  `db:"name"`
	Year   int64   `db:"year"`
	Plate  *string `db:"plate"`
}

func (s *CarStorage) SelectCarsFromDB(ctx context.Context, userID int64) ([]Car, error) {
	var cars []Car
	stmt := `
		select id, user_id, name, year, plate
		from car
		where user_id = ?
		order by year, name;
	`
	err := s.db.SelectContext(ctx, &cars, stmt, userID)
	return cars, err
}

func (s *CarStorage) GetCarFromDB(ctx context.Context, userID int64, carID int64) (Car, error) {
	var car Car
	stmt := `
		select id, user_id, name, year, plate
		from car
		where user_id = ? and id = ?;
	`
	err := s.db.GetContext(ctx, &car, stmt, userID, carID)
	return car, err
}

func (s *CarStorage) DeleteCarFromDB(ctx context.Context, userID int64, carID int64) (int64, error) {
	stmt := `delete from car where user_id = ? and id = ?;`
	res, err := s.db.ExecContext(ctx, stmt, userID, carID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *CarStorage) InsertCarIntoDB(ctx context.Context, car Car) (int64, error) {
	stmt := "insert into car (user_id, name, year, plate) values (?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, car.UserID, car.Name, car.Year, car.Plate)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
