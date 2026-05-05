-- +goose Up
-- +goose StatementBegin
create table users(
    id bigserial primary key not null,
    phone text not null unique,
    on_work boolean not null default false,
    verified boolean not null default false,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
