-- +goose Up
-- +goose StatementBegin
create table blob (
    id integer primary key,
    user_id integer not null,
    is_deleted boolean not null,
    uploaded_at integer not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table blob;
-- +goose StatementEnd
