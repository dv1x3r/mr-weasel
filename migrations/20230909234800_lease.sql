-- +goose Up
-- +goose StatementBegin
create table lease (
    id integer primary key,
    car_id integer not null references car(id) on delete cascade,
    timestamp integer not null,
    description text,
    cents integer not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table lease;
-- +goose StatementEnd
