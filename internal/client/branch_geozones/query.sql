-- name: GetByCoords :one
SELECT organization_branch_id, name
FROM branch_geozones
WHERE ST_Contains(
    zone,
    ST_SetSRID(ST_Point($1::numeric, $2::numeric), 4326) -- Передаем (Долгота, Широта)
);

-- name: GetByID :one
select *
from branch_geozones
where id = $1;