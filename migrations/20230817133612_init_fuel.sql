-- +goose Up
-- +goose StatementBegin
create table fuel (
    id integer primary key,
    car_id integer not null,
    timestamp integer not null,
    type text not null,
    volume integer not null,
    mileage integer not null,
    paid integer not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table fuel;
-- +goose StatementEnd
