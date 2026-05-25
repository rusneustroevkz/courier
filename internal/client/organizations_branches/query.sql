-- name: Create :one
insert into organization_branches(organization_id, name, address, latitude, longitude, phone)
values($1, $2, $3, $4, $5, $6)
returning id;

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

-- name: SetActivation :exec
update organization_branches
set is_active = $1
where organization_id = $2 and id = $3;

-- name: SetUserSelected :exec
update organization_branches
set user_selected = $1
where organization_id = $2 and id = $3;

-- name: GetCurrentSelected :one
select *
from organization_branches
where organization_id = $1 and user_selected = $2;

-- name: SetNullUserSelected :exec
update organization_branches
set user_selected = null
where organization_id = $1 and id = $2;