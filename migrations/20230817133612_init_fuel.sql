-- +goose Up
-- +goose StatementBegin
create table fuel (
    id integer primary key,
    car_id integer not null,
    timestamp integer not null,
    type text not null,
    cents integer not null
    milliliters integer not null,
    kilometers integer not null,
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table fuel;
-- +goose StatementEnd
