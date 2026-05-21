-- +goose Up
-- +goose StatementBegin
create table organization_branches (
    id bigserial primary key not null,
    organization_id bigint not null references organizations(id) on delete cascade,
    name text not null, -- Название (например: "Ресторан на Ленина", "Ресторан на Мира")
    address text not null, -- Полный адрес для курьера
    latitude numeric(9, 6), -- Координаты для карты и расчета дистанции
    longitude numeric(9, 6),
    phone text[], -- Телефон конкретной точки
    is_active boolean not null default true,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);

-- Индекс для быстрого поиска всех ресторанов одной компании
create index idx_organization_branches_org_id on organization_branches(organization_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table organization_branches;
-- +goose StatementEnd
