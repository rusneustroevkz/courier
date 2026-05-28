-- name: GetByID :one
select *
from orders
where id = $1;

-- name: GetPendingOrders :many
select *
from orders
where status = 'created';

-- name: AcceptOrder :exec
update orders
set courier_id = $1, status = $2, tg_courier_chat_id = $3, accepted_at = now()
where id = $4;