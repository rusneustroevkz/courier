-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE auth (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       TEXT NOT NULL,
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at  TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_refresh_tokens_token ON auth(token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table auth;
-- +goose StatementEnd
