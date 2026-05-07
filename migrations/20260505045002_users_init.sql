-- +goose Up
-- +goose StatementBegin
create type role_type as enum('courier','client','admin');
create type transport_type as enum('walk','bicycle','scooter','motorcycle','car','van','truck');

create table users(
    id bigserial primary key not null,
    tg_id bigint unique,
    full_name text,
    email text unique,
    phone text unique,
    role role_type not null,
    on_work boolean not null default false,
    verified boolean not null default false,
    rating numeric(3, 2) default 5.0,
    balance numeric(15, 2) default 0,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop type role_type;
drop type transport_type;
drop table users;
-- +goose StatementEnd
