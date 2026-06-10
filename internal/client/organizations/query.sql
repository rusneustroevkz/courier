-- name: GetByID :one
select *
from organizations
where id = $1;

-- name: GetByUserID :one
SELECT o.id, o.name, o.balance
FROM organizations o
         INNER JOIN users u ON o.id = u.organization_id
WHERE u.id = $1;