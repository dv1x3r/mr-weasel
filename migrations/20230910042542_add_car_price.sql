-- +goose Up
-- +goose StatementBegin
alter table car add column price integer;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table car drop column price;
-- +goose StatementEnd
