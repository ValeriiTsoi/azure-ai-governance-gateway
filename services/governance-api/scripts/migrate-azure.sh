#!/usr/bin/env bash

set -euo pipefail

RESOURCE_GROUP="${AZURE_RESOURCE_GROUP:-rg-aigov-demo}"
POSTGRES_SERVER="${POSTGRES_SERVER_NAME:-psql-aigov-0d60fe3d}"
POSTGRES_DATABASE="${POSTGRES_DATABASE_NAME:-aigov}"
POSTGRES_ADMIN="${POSTGRES_ADMIN_LOGIN:-aigovadmin}"

KEY_VAULT="${KEY_VAULT_NAME:-kv-aigov-0d60fe3d}"
PASSWORD_SECRET="${POSTGRES_PASSWORD_SECRET_NAME:-postgresql-admin-password}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


FIREWALL_RULE="migration-$(date +%Y%m%d%H%M%S)"
FIREWALL_CREATED=0

cleanup() {
  unset PGPASSWORD || true

  if [[ "${FIREWALL_CREATED}" == "1" ]]; then
    echo "Removing temporary PostgreSQL firewall rule..."

    az postgres flexible-server firewall-rule delete \
      --resource-group "${RESOURCE_GROUP}" \
      --server-name "${POSTGRES_SERVER}" \
      --name "${FIREWALL_RULE}" \
      --yes \
      --only-show-errors \
      >/dev/null || true
  fi
}

trap cleanup EXIT INT TERM


echo "Detecting current public IPv4 address..."

PUBLIC_IP="$(curl -4fsS https://api.ipify.org)"

if [[ ! "${PUBLIC_IP}" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
  echo "ERROR: unable to determine valid public IPv4 address."
  exit 1
fi

echo "Current public IP: ${PUBLIC_IP}"
echo "Creating temporary PostgreSQL firewall rule..."

az postgres flexible-server firewall-rule create \
  --resource-group "${RESOURCE_GROUP}" \
  --server-name "${POSTGRES_SERVER}" \
  --name "${FIREWALL_RULE}" \
  --start-ip-address "${PUBLIC_IP}" \
  --end-ip-address "${PUBLIC_IP}" \
  --only-show-errors \
  --output none

FIREWALL_CREATED=1

echo "Loading PostgreSQL password from Azure Key Vault..."

export PGPASSWORD="$(
  az keyvault secret show \
    --vault-name "${KEY_VAULT}" \
    --name "${PASSWORD_SECRET}" \
    --query value \
    --output tsv
)"

if [[ -z "${PGPASSWORD}" ]]; then
  echo "ERROR: PostgreSQL password is empty."
  exit 1
fi

PG_CONN="host=${POSTGRES_SERVER}.postgres.database.azure.com port=5432 dbname=${POSTGRES_DATABASE} user=${POSTGRES_ADMIN} sslmode=require"

echo "Waiting for PostgreSQL firewall rule to become effective..."

CONNECTED=0

for attempt in $(seq 1 30); do
  if docker run --rm \
      -e PGPASSWORD \
      postgres:16-alpine \
      psql "${PG_CONN}" \
      -Atqc 'SELECT 1;' \
      >/dev/null 2>&1
  then
    CONNECTED=1
    echo "PostgreSQL connection established."
    break
  fi

  printf 'Connection attempt %02d/30 failed; retrying in 10 seconds...\n' "${attempt}"
  sleep 10
done

if [[ "${CONNECTED}" != "1" ]]; then
  echo "ERROR: PostgreSQL could not be reached after 5 minutes."
  exit 1
fi

echo "Applying database migrations..."

export PG_CONN

"${SCRIPT_DIR}/run-migrations.sh"

echo
echo "Governance tables:"

docker run --rm \
  -e PGPASSWORD \
  postgres:16-alpine \
  psql "${PG_CONN}" \
  -P pager=off \
  -c "
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = 'public'
      AND tablename IN (
        'governance_requests',
        'policy_decisions',
        'model_routes',
        'usage_records',
        'schema_migrations'
      )
    ORDER BY tablename;
  "

echo
echo "Azure PostgreSQL migration completed."
