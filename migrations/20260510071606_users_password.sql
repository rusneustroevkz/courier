-- +goose Up
-- +goose StatementBegin
alter table users
add column password_hash text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
drop column password_hash;
-- +goose StatementEnd
