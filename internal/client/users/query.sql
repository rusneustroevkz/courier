-- name: GetByID :one
select *
from users
where id = $1;

-- name: GetByEmail :one
select *
from users
where email = $1;

-- name: CreateByEmail :exec
insert into users(email, role, password_hash, full_name)
values($1, $2, $3, $4);

-- name: ListOnWorkCourier :many
select *
from users
where on_work = true;