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

-- name: SetOnWork :exec
update users
set on_work = $1
where tg_id = $2;

-- name: SetShareLocation :exec
update users
set is_share_location = $1, share_location_ttl = $2, on_work = $3
where tg_id = $4;

-- name: ExpiredShareLocationList :many
select *
from users
where share_location_ttl < now();