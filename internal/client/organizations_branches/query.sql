-- name: Create :exec
insert into organization_branches(organization_id, name, address, latitude, longitude, phone)
values($1, $2, $3, $4, $5, $6);

-- name: List :many
select *
from organization_branches
where organization_id = $1;

-- name: GetByID :one
select *
from organization_branches
where organization_id = $1 and id = $2;

-- name: GetByName :one
select *
from organization_branches
where organization_id = $1 and name = $2;