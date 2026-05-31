-- +goose Up
-- +goose StatementBegin
alter table orders
    drop column courier_id,
    add column courier_id bigint references users(tg_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table orders
    drop column courier_id,
    add column courier_id bigint references users(id);
-- +goose StatementEnd
