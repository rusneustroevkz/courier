-- name: Create :one
insert into users(tg_id, full_name, email, phone, role)
values($1, $2, $3, $4, $5)
returning id;