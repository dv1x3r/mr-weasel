-- +goose Up
-- +goose StatementBegin
create table car_fuel (
    id int primary key,
    car_id int not null,
    timestamp int not null,
    fueltype text not null,
    liters real not null,
    totalcost int not null,
    odometer int not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table car_fuel;
-- +goose StatementEnd
