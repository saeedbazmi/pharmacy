-- name: ListOpsPharmacies :many
SELECT id, slug, name, site_domain, logo_url, status
FROM pharmacies
WHERE deleted_at IS NULL
ORDER BY name;

-- name: GetOpsPharmacy :one
SELECT id, slug, name, site_domain, logo_url, status
FROM pharmacies
WHERE id = $1
  AND deleted_at IS NULL;

-- name: InsertPharmacy :one
INSERT INTO pharmacies (slug, name, site_domain, logo_url, status)
VALUES ($1, $2, $3, $4, 'active')
RETURNING id;

-- name: UpdatePharmacy :exec
UPDATE pharmacies
SET name        = sqlc.arg(name),
    site_domain = sqlc.arg(site_domain),
    logo_url    = sqlc.arg(logo_url),
    status      = sqlc.arg(status)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: DisablePharmacy :exec
UPDATE pharmacies
SET status     = 'disabled',
    deleted_at = now()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ListOpsSources :many
SELECT ds.id,
       ds.pharmacy_id,
       p.name        AS pharmacy_name,
       p.slug        AS pharmacy_slug,
       ds.kind,
       ds.config,
       ds.schedule_interval::text AS schedule_interval,
       ds.is_enabled,
       ds.last_run_at,
       ds.last_status,
       ds.last_error
FROM data_sources ds
         JOIN pharmacies p ON p.id = ds.pharmacy_id
WHERE ds.deleted_at IS NULL
ORDER BY ds.id;

-- name: GetOpsSource :one
SELECT ds.id,
       ds.pharmacy_id,
       p.name        AS pharmacy_name,
       p.slug        AS pharmacy_slug,
       ds.kind,
       ds.config,
       ds.schedule_interval::text AS schedule_interval,
       ds.is_enabled,
       ds.last_run_at,
       ds.last_status,
       ds.last_error
FROM data_sources ds
         JOIN pharmacies p ON p.id = ds.pharmacy_id
WHERE ds.id = sqlc.arg(id)
  AND ds.deleted_at IS NULL;

-- name: InsertDataSource :one
INSERT INTO data_sources (pharmacy_id, kind, config, schedule_interval, is_enabled)
VALUES ($1, $2, $3, CAST(sqlc.arg(schedule_interval) AS text)::interval, TRUE)
RETURNING id;

-- name: UpdateDataSource :exec
UPDATE data_sources
SET kind              = sqlc.arg(kind),
    config            = sqlc.arg(config),
    schedule_interval = CAST(sqlc.arg(schedule_interval) AS text)::interval,
    is_enabled        = sqlc.arg(is_enabled)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: DisableDataSource :exec
UPDATE data_sources
SET is_enabled = FALSE,
    deleted_at = now()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ListSourceJobs :many
SELECT id, kind, status, attempts, max_attempts, run_at, last_error, created_at
FROM jobs
WHERE kind = 'sync_source'
  AND payload->>'source_id' = sqlc.arg(source_id)::text
ORDER BY created_at DESC
LIMIT 20;

-- name: GetLatestSyncRun :one
SELECT id, status, started_at, finished_at, ok_count, fail_count, error
FROM sync_runs
WHERE source_id = $1
ORDER BY started_at DESC
LIMIT 1;
