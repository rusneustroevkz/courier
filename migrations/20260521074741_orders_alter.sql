-- +goose Up
-- +goose StatementBegin

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS branch_id bigint REFERENCES organization_branches(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS courier_earnings numeric(15, 2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS delivery_distance_meters integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS tg_courier_chat_id bigint,
    ADD COLUMN IF NOT EXISTS accepted_at timestamp,
    ADD COLUMN IF NOT EXISTS picked_up_at timestamp,
    ADD COLUMN IF NOT EXISTS delivered_at timestamp,
    ADD COLUMN IF NOT EXISTS cancelled_at timestamp;

-- 3. Перестраиваем индексы для высокой производительности
DROP INDEX IF EXISTS idx_orders_courier_active; -- Удаляем старый индекс, если он был

CREATE INDEX IF NOT EXISTS idx_orders_organization_id ON orders(organization_id);
CREATE INDEX IF NOT EXISTS idx_orders_branch_id ON orders(branch_id);
-- Частичный индекс: ускоряет поиск только активных заказов курьера
CREATE INDEX IF NOT EXISTS idx_orders_courier_active_v2 ON orders(courier_id)
    WHERE status NOT IN ('delivered', 'cancelled');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Откат миграции в исходное состояние
DROP INDEX IF EXISTS idx_orders_courier_active_v2;
DROP INDEX IF EXISTS idx_orders_branch_id;
DROP INDEX IF EXISTS idx_orders_organization_id;

ALTER TABLE orders
DROP COLUMN IF EXISTS branch_id,
    DROP COLUMN IF EXISTS courier_earnings,
    DROP COLUMN IF EXISTS delivery_distance_meters,
    DROP COLUMN IF EXISTS tg_courier_chat_id,
    DROP COLUMN IF EXISTS accepted_at,
    DROP COLUMN IF EXISTS picked_up_at,
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS cancelled_at;

CREATE INDEX idx_orders_courier_active ON orders(courier_id);

-- +goose StatementEnd
