-- +goose Up
-- +goose StatementBegin
create table fuel (
    id int primary key,
    car_id int not null,
    timestamp int not null,
    type_fuel text not null,
    amount_ml int not null,
    amount_paid int not null,
    odometer int not null
) strict;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table fuel;
-- +goose StatementEnd
