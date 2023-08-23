-- +goose Up
-- +goose StatementBegin
create table holiday (
    id integer primary key,
    user_id integer not null,
    start integer not null,
    end integer not null,
    days integer not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table holiday;
-- +goose StatementEnd
