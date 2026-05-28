-- +goose Up
-- +goose StatementBegin
alter table orders
    drop column tg_client_chat_id,
    drop column tg_live_message_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table orders
    add column tg_client_chat_id bigint,
    add column tg_live_message_id bigint;
-- +goose StatementEnd
