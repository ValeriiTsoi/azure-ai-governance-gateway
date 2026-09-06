#!/usr/bin/env bash
set -euo pipefail

# Open temporary exact-IP PostgreSQL access for the demo.

RG="${AZURE_RESOURCE_GROUP:-rg-aigov-demo}"
SERVER="${POSTGRES_SERVER_NAME:-psql-aigov-0d60fe3d}"
DB="${POSTGRES_DATABASE_NAME:-aigov}"
DB_USER="${POSTGRES_ADMIN_LOGIN:-aigovadmin}"
KV="${KEY_VAULT_NAME:-kv-aigov-0d60fe3d}"
PASSWORD_SECRET="${POSTGRES_PASSWORD_SECRET_NAME:-postgresql-admin-password}"
STATE_FILE="${TMPDIR:-/tmp}/aigov-demo-db-access.env"

for cmd in az curl docker; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "ERROR: required command not found: $cmd" >&2
    exit 1
  fi
done

az account show --only-show-errors >/dev/null

# Refuse to stack another demo firewall rule.
if [[ -f "$STATE_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$STATE_FILE"

  if [[ -n "${FIREWALL_RULE:-}" ]]; then
    EXISTING_COUNT="$(
      az postgres flexible-server firewall-rule list \
        --resource-group "$RG" \
        --server-name "$SERVER" \
        --query "[?name=='$FIREWALL_RULE'] | length(@)" \
        -o tsv
    )"

    if [[ "$EXISTING_COUNT" != "0" ]]; then
      echo "ERROR: demo DB access is already open: $FIREWALL_RULE" >&2
      echo "Run ./scripts/demo-db-access-close.sh first." >&2
      exit 1
    fi
  fi

  rm -f "$STATE_FILE"
fi

PUBLIC_IP="$(curl -4fsS https://api.ipify.org)"

if [[ -z "$PUBLIC_IP" || "$PUBLIC_IP" =~ [^0-9.] ]]; then
  echo "ERROR: could not determine a valid public IPv4 address." >&2
  exit 1
fi

FIREWALL_RULE="demo-audit-$(date +%Y%m%d%H%M%S)"
RULE_CREATED=0

cleanup_on_error() {
  rc=$?
  trap - EXIT
  unset PGPASSWORD || true

  if [[ "$rc" -ne 0 && "$RULE_CREATED" -eq 1 ]]; then
    echo "Cleaning up failed demo access..."
    az postgres flexible-server firewall-rule delete \
      --resource-group "$RG" \
      --server-name "$SERVER" \
      --name "$FIREWALL_RULE" \
      --yes \
      --only-show-errors \
      >/dev/null 2>&1 || true

    rm -f "$STATE_FILE"
  fi

  exit "$rc"
}

trap cleanup_on_error EXIT

echo "Public IP: $PUBLIC_IP"
echo "Creating temporary firewall rule: $FIREWALL_RULE"

az postgres flexible-server firewall-rule create \
  --resource-group "$RG" \
  --server-name "$SERVER" \
  --name "$FIREWALL_RULE" \
  --start-ip-address "$PUBLIC_IP" \
  --end-ip-address "$PUBLIC_IP" \
  --only-show-errors \
  --output none

RULE_CREATED=1

umask 077
printf 'FIREWALL_RULE=%s\nPUBLIC_IP=%s\n' \
  "$FIREWALL_RULE" \
  "$PUBLIC_IP" \
  > "$STATE_FILE"

# Load DB password without printing it.
export PGPASSWORD="$(
  az keyvault secret show \
    --vault-name "$KV" \
    --name "$PASSWORD_SECRET" \
    --query value \
    -o tsv
)"

# Wait for Azure firewall propagation.
READY=0

for i in {1..24}; do
  if docker run --rm \
      -e PGPASSWORD \
      postgres:16-alpine \
      psql \
        "host=${SERVER}.postgres.database.azure.com port=5432 dbname=${DB} user=${DB_USER} sslmode=require" \
        -Atqc 'SELECT 1' \
        >/dev/null 2>&1
  then
    READY=1
    break
  fi

  echo "Waiting for firewall propagation... ($i/24)"
  sleep 5
done

unset PGPASSWORD

if [[ "$READY" -ne 1 ]]; then
  echo "ERROR: PostgreSQL did not become reachable." >&2
  exit 1
fi

trap - EXIT

echo
echo "Demo DB access: READY"
echo "Firewall rule: $FIREWALL_RULE"
echo "Run ./scripts/demo-db-access-close.sh after the demo."