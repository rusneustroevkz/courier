-- name: PayOrder :exec
update organizations
set balance = $1
where id = $2;

-- name: GetByID :one
select *
from organizations
where id = $1;