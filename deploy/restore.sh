#!/usr/bin/env bash
# Restore a gzipped SQL dump onto a Postgres instance.
# Usage: BACKUP=deploy/backups/pharmacy-YYYYMMDD.sql.gz ./deploy/restore.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP="${BACKUP:-}"
if [[ -z "$BACKUP" || ! -f "$BACKUP" ]]; then
  echo "set BACKUP to an existing .sql.gz file" >&2
  exit 1
fi

if [[ "${CONFIRM:-}" != "yes" ]]; then
  echo "this replaces the target database. re-run with CONFIRM=yes" >&2
  exit 1
fi

if [[ -n "${DATABASE_URL:-}" ]]; then
  gzip -dc "$BACKUP" | psql "$DATABASE_URL"
else
  gzip -dc "$BACKUP" | docker compose -f "$ROOT/docker-compose.yml" exec -T db \
    psql -U "${POSTGRES_USER:-pharmacy}" "${POSTGRES_DB:-pharmacy}"
fi

echo "restore applied: $BACKUP"
