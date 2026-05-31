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

-- name: GetCourierActiveOrder :one
select *
from orders
where courier_id = $1 and status not in ('created','delivered','cancelled')
limit 1;

-- name: DoneOrder :exec
update orders
set status = 'delivered'
where id = $1;