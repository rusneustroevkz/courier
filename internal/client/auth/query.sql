-- name: GetActiveRefreshToken :one
SELECT * FROM auth
WHERE token = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();

-- name: Save :exec
insert into auth (user_id, token, expires_at, revoked_at)
values ($3, $1, $2, NULL)
on conflict (user_id) do update
set
    token = EXCLUDED.token,
    expires_at = EXCLUDED.expires_at,
    revoked_at = NULL;

-- name: Logout :exec
update auth
set revoked_at = $1
where user_id = $2;

-- name: RevokeToken :exec
UPDATE auth
SET revoked_at = CURRENT_TIMESTAMP
WHERE token = $1 AND revoked_at IS NULL;