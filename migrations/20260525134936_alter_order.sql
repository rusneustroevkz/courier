-- +goose Up
-- +goose StatementBegin
alter table orders
drop column price;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table orders
add column price numeric(15,2);
-- +goose StatementEnd
