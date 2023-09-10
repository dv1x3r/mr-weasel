package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
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
	Price  sql.NullInt64  `db:"price"`
}

type CarDetails struct {
	CarBase
	Kilometers int64 `db:"kilometers"`
}

func (s *CarStorage) SelectCarsFromDB(ctx context.Context, userID int64) ([]CarDetails, error) {
	var cars []CarDetails
	stmt := `
		select id, user_id, name, year, plate, price
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
		select 
			c.id
			,c.user_id
			,c.name
			,c.year
			,c.plate
			,c.price
			,coalesce(f.kilometers, 0) as kilometers
		from car c
		left join (
			select
				car_id
				,max(kilometers) as kilometers
			from fuel
			group by car_id
		) f on f.car_id = c.id
		where c.user_id = ? and c.id = ?;
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
	stmt := "insert into car (user_id, name, year, plate, price) values (?,?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, car.UserID, car.Name, car.Year, car.Plate, car.Price)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *CarStorage) UpdateCarInDB(ctx context.Context, car CarBase) (int64, error) {
	stmt := "update car set user_id = ?, name = ?, year = ?, plate = ?, price = ? where id = ?;"
	res, err := s.db.ExecContext(ctx, stmt, car.UserID, car.Name, car.Year, car.Plate, car.Price, car.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
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

func (f *FuelBase) GetTimestamp() string {
	return time.Unix(f.Timestamp, 0).UTC().Format("Monday, 02 January 2006")
}

func (f *FuelBase) GetLiters() float64 {
	return float64(f.Milliliters) / 1000
}

func (f *FuelBase) GetEuro() float64 {
	return float64(f.Cents) / 100
}

func (f *FuelBase) GetEurPerLiter() float64 {
	return f.GetEuro() / f.GetLiters()
}

func (f *FuelDetails) GetEurPerKilometer() float64 {
	return f.GetEuro() / float64(f.KilometersR)
}

func (f *FuelDetails) GetLitersPerKilometer() float64 {
	return (f.GetLiters() / float64(f.KilometersR)) * 100
}

func (s *CarStorage) GetFuelFromDB(ctx context.Context, userID int64, carID int64, offset int64) (FuelDetails, error) {
	var fuel FuelDetails
	stmt := `
		select
			f.id
			,f.car_id
			,f.timestamp
			,f.type
			,f.milliliters
			,f.kilometers
			,f.cents
			,coalesce(f.kilometers - lag(f.kilometers) over (order by f.timestamp, f.id), f.kilometers) as kilometersr
			,count(*) over (partition by car_id) as countrows
		from fuel f
		join car c on c.id = f.car_id 
		where c.user_id = ? and c.id = ?
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
			and car_id in (select id from car where user_id = ?);
	`
	res, err := s.db.ExecContext(ctx, stmt, fuelID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type ServiceBase struct {
	ID          int64  `db:"id"`
	CarID       int64  `db:"car_id"`
	Timestamp   int64  `db:"timestamp"`
	Description string `db:"description"`
	Cents       int64  `db:"cents"`
}

type ServiceDetails struct {
	ServiceBase
	CountRows int64 `db:"countrows"`
}

func (s *ServiceBase) GetEuro() float64 {
	return float64(s.Cents) / 100
}

func (s *ServiceBase) GetTimestamp() string {
	return time.Unix(s.Timestamp, 0).UTC().Format("Monday, 02 January 2006")
}

func (s *CarStorage) GetServiceFromDB(ctx context.Context, userID int64, carID int64, offset int64) (ServiceDetails, error) {
	var service ServiceDetails
	stmt := `
		select
			s.id
			,s.car_id
			,s.timestamp
			,s.description
			,s.cents
			,count(*) over (partition by car_id) as countrows
		from service s
		join car c on c.id = s.car_id 
		where c.user_id = ? and c.id = ?
		order by s.timestamp desc, s.id desc
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &service, stmt, userID, carID, offset)
	return service, err
}

func (s *CarStorage) InsertServiceIntoDB(ctx context.Context, service ServiceBase) (int64, error) {
	stmt := "insert into service (car_id, timestamp, description, cents) values (?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, service.CarID, service.Timestamp, service.Description, service.Cents)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *CarStorage) DeleteServiceFromDB(ctx context.Context, userID int64, serviceID int64) (int64, error) {
	stmt := `
		delete from service where id = ?
			and car_id in (select id from car where user_id = ?);
	`
	res, err := s.db.ExecContext(ctx, stmt, serviceID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type LeaseBase struct {
	ID          int64          `db:"id"`
	CarID       int64          `db:"car_id"`
	Timestamp   int64          `db:"timestamp"`
	Description sql.NullString `db:"description"`
	Cents       int64          `db:"cents"`
}

type LeaseDetails struct {
	LeaseBase
	CountRows int64 `db:"countrows"`
}

func (s *LeaseBase) GetEuro() float64 {
	return float64(s.Cents) / 100
}

func (s *LeaseBase) GetTimestamp() string {
	return time.Unix(s.Timestamp, 0).UTC().Format("Monday, 02 January 2006")
}

func (s *CarStorage) GetLeaseFromDB(ctx context.Context, userID int64, carID int64, offset int64) (LeaseDetails, error) {
	var lease LeaseDetails
	stmt := `
		select
			l.id
			,l.car_id
			,l.timestamp
			,l.description
			,l.cents
			,count(*) over (partition by car_id) as countrows
		from lease l
		join car c on c.id = l.car_id 
		where c.user_id = ? and c.id = ?
		order by l.timestamp desc, l.id desc
		limit 1 offset ?;
	`
	err := s.db.GetContext(ctx, &lease, stmt, userID, carID, offset)
	return lease, err
}

func (s *CarStorage) InsertLeaseIntoDB(ctx context.Context, lease LeaseBase) (int64, error) {
	stmt := "insert into lease (car_id, timestamp, description, cents) values (?,?,?,?);"
	res, err := s.db.ExecContext(ctx, stmt, lease.CarID, lease.Timestamp, lease.Description, lease.Cents)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *CarStorage) DeleteLeaseFromDB(ctx context.Context, userID int64, leaseID int64) (int64, error) {
	stmt := `
		delete from lease where id = ?
			and car_id in (select id from car where user_id = ?);
	`
	res, err := s.db.ExecContext(ctx, stmt, leaseID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
