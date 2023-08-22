-- +goose Up
-- +goose StatementBegin
create table service (
    id integer primary key,
    car_id integer not null references car(id) on delete cascade,
    timestamp integer not null,
    description text not null,
    cents integer not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table service;
-- +goose StatementEnd
