-- +goose Up
-- +goose StatementBegin
create table car (
    id integer primary key,
    user_id integer not null,
    name text not null,
    year int not null,
    plate text
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table car;
-- +goose StatementEnd
