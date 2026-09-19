-- name: EnqueueJob :one
INSERT INTO jobs (kind, payload, max_attempts, run_at)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: HasOpenJob :one
SELECT EXISTS (
    SELECT 1
    FROM jobs
    WHERE kind = $1
      AND payload->>'source_id' = sqlc.arg(source_id)::text
      AND status IN ('pending', 'retry', 'running')
) AS exists;

-- name: ClaimJob :one
UPDATE jobs
SET status = 'running',
    attempts = attempts + 1,
    updated_at = now()
WHERE id = (
    SELECT id
    FROM jobs
    WHERE status IN ('pending', 'retry')
      AND run_at <= now()
    ORDER BY run_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING id, kind, payload, status, attempts, max_attempts, run_at, last_error;

-- name: CompleteJob :exec
UPDATE jobs
SET status = 'succeeded',
    last_error = NULL,
    updated_at = now()
WHERE id = $1;

-- name: FailJob :exec
UPDATE jobs
SET status = sqlc.arg(status),
    run_at = sqlc.arg(run_at),
    last_error = sqlc.arg(last_error),
    updated_at = now()
WHERE id = sqlc.arg(id);

-- name: HasOpenJobByKind :one
SELECT EXISTS (
    SELECT 1
    FROM jobs
    WHERE kind = $1
      AND status IN ('pending', 'retry', 'running')
) AS exists;
