-- +goose Up
-- +goose StatementBegin
create table user (
    id integer primary key,
    telegram_id integer unique,
    name text
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table user;
-- +goose StatementEnd
