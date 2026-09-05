#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
MIGRATION_DIR="${SERVICE_DIR}/migrations"

: "${PG_CONN:?PG_CONN must be set}"
: "${PGPASSWORD:?PGPASSWORD must be set}"

PSQL_IMAGE="${PSQL_IMAGE:-postgres:16-alpine}"

psql_exec() {
  docker run --rm -i \
    -e PGPASSWORD \
    "${PSQL_IMAGE}" \
    psql "${PG_CONN}" "$@"
}

echo "Checking migration tracking state..."

TRACKING_EXISTS="$(
  psql_exec     -Atqc "
      SELECT to_regclass('public.schema_migrations') IS NOT NULL;
    "
)"

CORE_TABLE_COUNT="$(
  psql_exec     -Atqc "
      SELECT count(*)
      FROM pg_tables
      WHERE schemaname = 'public'
        AND tablename IN (
          'governance_requests',
          'policy_decisions',
          'model_routes',
          'usage_records'
        );
    "
)"

if [[ "${TRACKING_EXISTS}" != "t" ]] &&
   [[ "${CORE_TABLE_COUNT}" != "0" ]]; then
  echo "ERROR: database schema exists but schema_migrations is missing."
  echo "Refusing to guess migration history."
  echo "Core tables found: ${CORE_TABLE_COUNT}"
  exit 1
fi

echo "Creating migration tracking table if necessary..."

psql_exec \
  -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT PRIMARY KEY,
    name        TEXT NOT NULL,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL

shopt -s nullglob

MIGRATION_FILES=(
  "${MIGRATION_DIR}"/*.up.sql
)

if [[ "${#MIGRATION_FILES[@]}" -eq 0 ]]; then
  echo "ERROR: no migration files found in ${MIGRATION_DIR}"
  exit 1
fi

for migration_file in "${MIGRATION_FILES[@]}"; do
  filename="$(basename "${migration_file}")"

  if [[ ! "${filename}" =~ ^([0-9]{6})_([A-Za-z0-9_]+)\.up\.sql$ ]]; then
    echo "ERROR: invalid migration filename: ${filename}"
    exit 1
  fi

  version_text="${BASH_REMATCH[1]}"
  migration_name="${BASH_REMATCH[2]}"
  version=$((10#${version_text}))

  existing_name="$(
    psql_exec \
      -Atqc "
        SELECT name
        FROM schema_migrations
        WHERE version = ${version};
      "
  )"

  if [[ -n "${existing_name}" ]]; then
    if [[ "${existing_name}" != "${migration_name}" ]]; then
      echo "ERROR: migration version ${version_text} name mismatch."
      echo "Database: ${existing_name}"
      echo "File:     ${migration_name}"
      exit 1
    fi

    echo "Migration ${version_text}_${migration_name} already applied."
    continue
  fi

  echo "Applying migration ${version_text}_${migration_name}..."

  sql_name="${migration_name//\'/\'\'}"

  {
    echo "BEGIN;"
    cat "${migration_file}"
    echo
    echo "INSERT INTO schema_migrations (version, name)"
    echo "VALUES (${version}, '${sql_name}');"
    echo "COMMIT;"
  } | psql_exec \
        -v ON_ERROR_STOP=1

  echo "Migration ${version_text}_${migration_name} applied successfully."
done

echo
echo "Migration status:"

psql_exec \
  -P pager=off \
  -c "
    SELECT
      version,
      name,
      applied_at
    FROM schema_migrations
    ORDER BY version;
  "

echo
echo "All migrations applied successfully."
