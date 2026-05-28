-- +goose Up
-- +goose StatementBegin
alter table users
    add column is_share_location boolean not null default false,
    add column share_location_ttl timestamptz;

CREATE INDEX idx_users_share_location_ttl ON users (share_location_ttl);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
    drop column is_share_location,
    drop column share_location_ttl;
-- +goose StatementEnd
