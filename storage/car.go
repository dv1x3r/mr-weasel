package storage

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

type CarStorage struct {
	db *sqlx.DB
}

func NewCarStorage(db *sqlx.DB) *CarStorage {
	return &CarStorage{db: db}
}

type CarBase struct {
	ID     int64          `db:"id"`
	UserID int64          `db:"user_id"`
	Name   string         `db:"name"`
	Year   int64          `db:"year"`
	Plate  sql.NullString `db:"plate"`
}

type CarDetails struct {
	CarBase
}

type FuelBase struct {
	ID          int64  `db:"id"`
	CarID       int64  `db:"car_id"`
	Timestamp   int64  `db:"timestamp"`
	Type        string `db:"type"`
	Cents       int64  `db:"cents"`
	Milliliters int64  `db:"milliliters"`
	Kilometers  int64  `db:"kilometers"`
}

type FuelDetails struct {
	FuelBase
	KilometersR int64 `db:"kilometersr"`
	CountRows   int64 `db:"countrows"`
}

func (f *FuelDetails) GetLiters() float64 {
	return float64(f.Milliliters) / 1000
}

func (f *FuelDetails) GetEuro() float64 {
	return float64(f.Cents) / 100
}

func (f *FuelDetails) GetEurPerLiter() float64 {
	return f.GetEuro() / f.GetLiters()
}

func (f *FuelDetails) GetEurPerKilometer() float64 {
	return f.GetEuro() / float64(f.Kilometers)
}

func (f *FuelDetails) GetLitersPerKilometer() float64 {
	return float64(f.KilometersR) / f.GetLiters()
}

func (f *FuelDetails) GetTimestamp() time.Time {
	return time.Unix(f.Timestamp, 0)
}

func (s *CarStorage) SelectCarsFromDB(ctx context.Context, userID int64) ([]CarDetails, error) {
	var cars []CarDetails
	stmt := `
		select id, user_id, name, year, plate
		from car
		where user_id = ?
		order by year, name;
	`
	err := s.db.SelectContext(ctx, &cars, stmt, userID)
	return cars, err
}

func (s *CarStorage) GetCarFromDB(ctx context.Context, userID int64, carID int64) (CarDetails, error) {
	var car CarDetails
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

func (s *CarStorage) InsertCarIntoDB(ctx context.Context, car CarBase) (int64, error) {
	stmt := "insert into car (user_id, name, year, plate) values (?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, car.UserID, car.Name, car.Year, car.Plate)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *CarStorage) UpdateCarInDB(ctx context.Context, car CarBase) (int64, error) {
	stmt := "update car set user_id = ?, name = ?, year = ?, plate = ? where id = ?;"
	res, err := s.db.ExecContext(ctx, stmt, car.UserID, car.Name, car.Year, car.Plate, car.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *CarStorage) GetFuelFromDB(ctx context.Context, userID int64, carID int64, offset int) (FuelDetails, error) {
	var fuel FuelDetails
	stmt := `
		select
			f.id,
			f.car_id,
			f.timestamp,
			f.type,
			f.milliliters,
			f.kilometers,
			f.cents,
			coalesce(f.kilometers - lag(f.kilometers) over (order by f.id), f.kilometers) as kilometersr,
			count(*) over (partition by car_id) as countrows
		from fuel f
		join car c on c.id = f.car_id 
		where c.user_id = ? and f.car_id = ?
		order by f.timestamp desc, f.id desc
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &fuel, stmt, userID, carID, offset)
	return fuel, err
}

func (s *CarStorage) InsertFuelIntoDB(ctx context.Context, fuel FuelBase) (int64, error) {
	stmt := "insert into fuel (car_id, timestamp, type, milliliters, kilometers, cents) values (?,?,?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, fuel.CarID, fuel.Timestamp, fuel.Type, fuel.Milliliters, fuel.Kilometers, fuel.Cents)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *CarStorage) DeleteFuelFromDB(ctx context.Context, userID int64, fuelID int64) (int64, error) {
	stmt := `
		delete from fuel where id = ?
			and car_id in (
				select id
				from car
				where user_id = ?
			);
	`
	res, err := s.db.ExecContext(ctx, stmt, fuelID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
