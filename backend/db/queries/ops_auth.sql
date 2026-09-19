-- name: GetInternalUserByUsername :one
SELECT id, username, password_hash, role, is_active
FROM internal_users
WHERE username = $1
  AND deleted_at IS NULL;

-- name: GetInternalUserByID :one
SELECT id, username, role, is_active
FROM internal_users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: InsertInternalUser :one
INSERT INTO internal_users (username, password_hash, role, is_active)
VALUES ($1, $2, $3, TRUE)
RETURNING id;

-- name: InsertSession :exec
INSERT INTO internal_sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetSessionUser :one
SELECT u.id, u.username, u.role, u.is_active, s.expires_at
FROM internal_sessions s
         JOIN internal_users u ON u.id = s.user_id AND u.deleted_at IS NULL
WHERE s.token_hash = $1;

-- name: DeleteSession :exec
DELETE FROM internal_sessions
WHERE token_hash = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM internal_sessions
WHERE expires_at < now();

-- name: InsertLoginAttempt :exec
INSERT INTO login_attempts (username, ip)
VALUES ($1, $2);

-- name: CountRecentLoginAttempts :one
SELECT COUNT(*)::bigint AS n
FROM login_attempts
WHERE attempted_at >= sqlc.arg(since)
  AND (username = sqlc.arg(username) OR ip = sqlc.arg(ip));
