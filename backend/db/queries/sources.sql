-- name: ListDueSources :many
SELECT ds.id,
       ds.pharmacy_id,
       ds.kind,
       ds.config,
       ds.is_enabled
FROM data_sources ds
WHERE ds.deleted_at IS NULL
  AND ds.is_enabled
  AND (ds.last_run_at IS NULL
    OR ds.last_run_at + ds.schedule_interval <= now());

-- name: GetDataSource :one
SELECT ds.id,
       ds.pharmacy_id,
       ds.kind,
       ds.config,
       ds.is_enabled
FROM data_sources ds
WHERE ds.id = $1
  AND ds.deleted_at IS NULL;

-- name: MarkSourceRun :exec
UPDATE data_sources
SET last_run_at = now(),
    last_status = sqlc.arg(last_status),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: InsertSyncRun :one
INSERT INTO sync_runs (source_id, status)
VALUES ($1, 'running')
RETURNING id;

-- name: FinishSyncRun :exec
UPDATE sync_runs
SET finished_at = now(),
    status = sqlc.arg(status),
    ok_count = sqlc.arg(ok_count),
    fail_count = sqlc.arg(fail_count),
    error = sqlc.arg(error)
WHERE id = sqlc.arg(id);
