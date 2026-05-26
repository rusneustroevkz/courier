-- name: GetByID :one
select *
from orders
where id = $1 and organization_id = $2;

-- name: CreateOrder :one
insert into orders(organization_id, from_address, from_lat, from_lon, to_address, to_lat, to_lon, description)
values($1, $2, $3, $4, $5, $6, $7, $8)
returning id;

-- name: GetAll :many
select *
from orders
where organization_id = $1
offset $2
limit $3;

-- name: UpdateCourier :exec
update orders
set courier_id = $1, tg_client_chat_id = $2, tg_live_message_id = $3, updated_at = $4
where id = $5;