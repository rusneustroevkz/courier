-- name: Create :exec
insert into users(tg_id, full_name, email, phone, role)
values($1, $2, $3, $4, $5);

-- name: GetByTgID :one
select *
from users
where tg_id = $1;

-- name: UpdatePhoneByTgID :exec
update users
set phone = $1
where tg_id = $2;