-- +goose Up
-- +goose StatementBegin
create table blob (
    id integer primary key,
    user_id integer not null,
    file_id text not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table blob;
-- +goose StatementEnd
