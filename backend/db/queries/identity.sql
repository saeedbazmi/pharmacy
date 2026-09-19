-- name: GetUserByPhone :one
SELECT id, phone, role
FROM users
WHERE phone = $1
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, phone, role
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: InsertUser :one
INSERT INTO users (phone)
VALUES ($1)
ON CONFLICT (phone) WHERE deleted_at IS NULL
DO UPDATE SET updated_at = now()
RETURNING id, phone, role;

-- name: InsertOTPCode :one
INSERT INTO otp_codes (phone, code_hash, expires_at, ip)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetLatestOTP :one
SELECT id, phone, code_hash, expires_at, consumed_at, attempt_count, created_at
FROM otp_codes
WHERE phone = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ConsumeOTP :execrows
UPDATE otp_codes
SET consumed_at = now()
WHERE id = $1
  AND consumed_at IS NULL;

-- name: IncrementOTPAttempts :exec
UPDATE otp_codes
SET attempt_count = attempt_count + 1
WHERE id = $1;

-- name: InsertOTPSend :exec
INSERT INTO otp_sends (phone, ip)
VALUES ($1, $2);

-- name: CountOTPSendsByPhone :one
SELECT COUNT(*)::bigint
FROM otp_sends
WHERE phone = $1
  AND created_at >= $2;

-- name: CountOTPSendsByIP :one
SELECT COUNT(*)::bigint
FROM otp_sends
WHERE ip = $1
  AND created_at >= $2;

-- name: InsertUserSession :exec
INSERT INTO user_sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetUserBySessionHash :one
SELECT u.id, u.phone, u.role, s.expires_at
FROM user_sessions s
         JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL
WHERE s.token_hash = $1
  AND s.expires_at > now();

-- name: DeleteUserSession :exec
DELETE FROM user_sessions
WHERE token_hash = $1;

-- name: ListUserSearches :many
SELECT id, queried_at, query_normalized, result_count
FROM search_queries
WHERE user_id = $1
ORDER BY queried_at DESC
LIMIT 50;

-- name: DeleteUserSearches :exec
DELETE FROM search_queries
WHERE user_id = $1;
