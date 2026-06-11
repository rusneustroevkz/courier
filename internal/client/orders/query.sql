-- name: GetByID :one
select *
from orders
where id = $1 and organization_id = $2;

-- name: CreateOrder :one
insert into orders(organization_id, from_address, from_lat, from_lon, to_address, to_lat, to_lon, description, price)
values($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning id;

-- name: GetAll :many
select *
from orders
where organization_id = $1
offset $2
limit $3;

-- name: Update :exec
update orders
set to_address = $1, to_lat = $2, to_lon = $3, description = $4, updated_at = now()
where id = $5 and organization_id = $6;

-- name: CancelOrder :exec
update orders
set status = $1, updated_at = now()
where id = $2 and organization_id = $3 and status not in ('delivered','cancelled');