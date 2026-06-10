-- 1. Включаем расширение PostGIS (если еще не включено)
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


INSERT INTO branch_geozones (organization_branch_id, name, zone)
VALUES (
7,
'Центральный район',
ST_GeomFromGeoJSON('{"type":"Polygon","coordinates":[[[129.7157074,62.0372632],[129.7422721,62.0472466],[129.7459288,62.0314128],[129.7218377,62.0222314],[129.7102224,62.0233414],[129.6993599,62.0284367],[129.709171,62.0338382],[129.7157074,62.0372632]]]}}')
);