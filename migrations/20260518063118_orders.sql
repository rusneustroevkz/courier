-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM (
    'created',     -- Создан клиентом, ищет курьера
    'accepted',    -- Курьер принял заказ и едет на точку А
    'picked_up',   -- Курьер забрал груз, едет на точку Б (включен трекинг)
    'delivered',   -- Успешно доставлен
    'cancelled'    -- Отменен
);
create table orders (
    id bigserial primary key not null,
    description text,

    -- Кто заказал и кто везет
    organization_id bigint not null references organizations(id), -- Организация, которая заказала
    courier_id bigint references users(id), -- NULL, пока заказ в статусе 'created'

    -- Статус
    status order_status not null default 'created',

    -- Адреса и координаты (для маршрута на бэкенде)
    from_address text not null,
    from_lat numeric(9, 6) not null,
    from_lon numeric(9, 6) not null,

    to_address text not null,
    to_lat numeric(9, 6) not null,
    to_lon numeric(9, 6) not null,

    -- Финансы
    price numeric(15, 2) not null default 0.00,

    -- Данные для Telegram Live Location (Критично для вашей логики!)
    tg_client_chat_id bigint, -- Чат с клиентом, куда бот кинул карту
    tg_live_message_id bigint, -- ID сообщения с картой, которое бэкенд будет постоянно обновлять

    -- Таймстампы
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
create index idx_orders_courier_active on orders(courier_id);
create index idx_orders_status on orders(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table orders;
drop type order_status;
-- +goose StatementEnd
