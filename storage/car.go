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

type Fuel struct {
	ID         int64  `db:"id"`
	CarID      int64  `db:"car_id"`
	Timestamp  int64  `db:"timestamp"`
	TypeFuel   string `db:"type_fuel"`
	AmountML   int64  `db:"amount_ml"`
	AmountPaid int64  `db:"amount_paid"`
	Odometer   int64  `db:"odometer"`
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

func (s *CarStorage) UpdateCarInDB(ctx context.Context, car Car) (int64, error) {
	stmt := "update car set user_id = ?, name = ?, year = ?, plate = ? where id = ?;"
	res, err := s.db.ExecContext(ctx, stmt, car.UserID, car.Name, car.Year, car.Plate, car.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *CarStorage) GetFuelFromDB(ctx context.Context, userID int64, carID int64, offset int) (Fuel, error) {
	var fuel Fuel
	stmt := `
		select f.id, f.car_id, f.timestamp, f.type_fuel, f.amount_ml, f.amount_paid, f.odometer
		from fuel f
		join car c on c.id = f.car_id 
		where c.user_id = ? and f.car_id = ?
		order by f.timestamp desc
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &fuel, stmt, userID, carID, offset)
	return fuel, err
}

func (s *CarStorage) InsertFuelIntoDB(ctx context.Context, fuel Fuel) (int64, error) {
	stmt := "insert into fuel (car_id, timestamp, type_fuel, amount_ml, amount_paid, odometer) values (?,?,?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, fuel.CarID, fuel.Timestamp, fuel.TypeFuel, fuel.AmountML, fuel.AmountPaid, fuel.Odometer)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
