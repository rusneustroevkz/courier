-- name: GetByID :one
select *
from users
where id = $1;

-- name: List :many
select *
from users
limit $1
offset $2;