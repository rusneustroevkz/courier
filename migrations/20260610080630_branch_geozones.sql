-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS postgis;

-- 2. Создаем таблицу геозон
CREATE TABLE branch_geozones (
    id SERIAL PRIMARY KEY,
    organization_branch_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    zone GEOMETRY(Polygon, 4326),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Создаем пространственный индекс для быстрого поиска
CREATE INDEX idx_branch_geozones_spatial ON branch_geozones USING GIST (zone);

-- 4. Создаем обычный индекс для поиска по филиалу
CREATE INDEX idx_branch_geozones_branch_id ON branch_geozones (organization_branch_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop EXTENSION postgis;
 drop table branch_geozones;
-- +goose StatementEnd
