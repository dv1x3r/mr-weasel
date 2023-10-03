-- +goose Up
-- +goose StatementBegin
create table rvc_model (
    id integer primary key,
    user_id integer not null,
    name text not null,
    dataset_folder text,
    model_file text,
    index_file text,
    is_trained integer not null,
    is_deleted integer not null,
    created_at integer not null,
    updated_at integer not null
) strict;

create table rvc_access (
    id integer primary key,
    model_id integer not null references rvc_model(id),
    user_id integer not null,
    user_name text not null
) strict;

create table rvc_experiment (
    id integer primary key,
    user_id integer not null,
    model_id integer references rvc_model(id),
    audio_id text,
    enable_uvr integer not null,
    transpose integer not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table rvc_model;
drop table rvc_access;
drop table rvc_experiment;
-- +goose StatementEnd
