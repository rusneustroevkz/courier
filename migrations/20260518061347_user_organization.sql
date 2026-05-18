-- +goose Up
-- +goose StatementBegin
alter table users add column organization_id bigint references organizations(id) on delete set null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users drop column organization_id;
-- +goose StatementEnd
