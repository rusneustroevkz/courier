-- +goose Up
-- +goose StatementBegin
alter table organization_branches
    add column user_selected bigint;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table organization_branches
    drop column user_selected;
-- +goose StatementEnd
