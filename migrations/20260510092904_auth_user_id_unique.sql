-- +goose Up
-- +goose StatementBegin
ALTER TABLE auth ADD CONSTRAINT auth_user_id_unique UNIQUE (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table auth drop constraint auth_user_id_unique;
-- +goose StatementEnd
