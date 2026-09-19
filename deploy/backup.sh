#!/usr/bin/env bash
# Daily database dump. Intended for the production host (cron).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEST_DIR="${BACKUP_DIR:-$ROOT/deploy/backups}"
mkdir -p "$DEST_DIR"
FILE="$DEST_DIR/pharmacy-$STAMP.sql.gz"

if [[ -n "${DATABASE_URL:-}" ]]; then
  pg_dump --no-owner --format=plain "$DATABASE_URL" | gzip -9 > "$FILE"
else
  docker compose -f "$ROOT/docker-compose.yml" exec -T db \
    pg_dump -U "${POSTGRES_USER:-pharmacy}" --no-owner --format=plain "${POSTGRES_DB:-pharmacy}" \
    | gzip -9 > "$FILE"
fi

# Keep 14 days.
find "$DEST_DIR" -name 'pharmacy-*.sql.gz' -mtime +14 -delete || true
echo "backup written: $FILE"
