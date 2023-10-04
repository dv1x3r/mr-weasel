-- +goose Up
-- +goose StatementBegin
create table rvc_model (
    id integer primary key,
    user_id integer not null,
    name text not null,
    dataset_folder text,
    model_file text,
    index_file text
) strict;

create table rvc_access (
    id integer primary key,
    model_id integer not null references rvc_model(id) on delete cascade,
    user_id integer not null
) strict;

create table rvc_experiment (
    id integer primary key,
    user_id integer not null,
    model_id integer references rvc_model(id) on delete set null,
    audio_source_id text,
    audio_voice_file text,
    audio_music_file text,
    separate_uvr integer not null default 0,
    transpose integer not null default 0
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table rvc_model;
drop table rvc_access;
drop table rvc_experiment;
-- +goose StatementEnd
