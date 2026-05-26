-- name: GetByID :one
select *
from orders
where id = $1;

-- name: GetPendingOrders :many
select *
from orders
where status = 'created';