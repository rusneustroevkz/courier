-- +goose Up
-- +goose StatementBegin
create table organizations (
    id bigserial primary key not null,
    name text not null,
    inn text unique,
    balance numeric(15, 2) not null default 0.00,
    is_active boolean not null default false,
    verified boolean not null default false,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table organizations;
-- +goose StatementEnd
