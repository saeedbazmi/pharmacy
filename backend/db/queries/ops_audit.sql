-- name: InsertAuditLog :exec
INSERT INTO audit_logs (actor_id, actor_name, entity, entity_id, action, before_data, after_data)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditLogs :many
SELECT id, actor_id, actor_name, entity, entity_id, action, before_data, after_data, created_at
FROM audit_logs
WHERE (sqlc.arg(actor_id)::bigint = 0 OR actor_id = sqlc.arg(actor_id))
  AND created_at >= sqlc.arg(since)
  AND created_at < sqlc.arg(until)
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAuditLogs :one
SELECT COUNT(*)::bigint AS n
FROM audit_logs
WHERE (sqlc.arg(actor_id)::bigint = 0 OR actor_id = sqlc.arg(actor_id))
  AND created_at >= sqlc.arg(since)
  AND created_at < sqlc.arg(until);
