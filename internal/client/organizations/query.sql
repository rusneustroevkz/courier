-- name: GetByID :one
select *
from organizations
where id = $1;