#!/usr/bin/env bash
set -euo pipefail

# Remove temporary demo PostgreSQL access.

RG="${AZURE_RESOURCE_GROUP:-rg-aigov-demo}"
SERVER="${POSTGRES_SERVER_NAME:-psql-aigov-0d60fe3d}"
STATE_FILE="${TMPDIR:-/tmp}/aigov-demo-db-access.env"

unset PGPASSWORD || true

if [[ ! -f "$STATE_FILE" ]]; then
  echo "Demo DB access: no active state found."
  exit 0
fi

# shellcheck disable=SC1090
source "$STATE_FILE"

if [[ -z "${FIREWALL_RULE:-}" ]]; then
  echo "ERROR: firewall rule is missing from state file." >&2
  exit 1
fi

case "$FIREWALL_RULE" in
  demo-audit-*)
    ;;
  *)
    echo "ERROR: refusing to delete unexpected rule: $FIREWALL_RULE" >&2
    exit 1
    ;;
esac

echo "Removing temporary firewall rule: $FIREWALL_RULE"

az postgres flexible-server firewall-rule delete \
  --resource-group "$RG" \
  --server-name "$SERVER" \
  --name "$FIREWALL_RULE" \
  --yes \
  --only-show-errors

# Verify Azure control-plane cleanup.
COUNT=""

for i in {1..15}; do
  COUNT="$(
    az postgres flexible-server firewall-rule list \
      --resource-group "$RG" \
      --server-name "$SERVER" \
      --query "[?name=='$FIREWALL_RULE'] | length(@)" \
      -o tsv
  )"

  if [[ "$COUNT" == "0" ]]; then
    break
  fi

  echo "Waiting for firewall cleanup... ($i/15)"
  sleep 2
done

if [[ "$COUNT" != "0" ]]; then
  echo "ERROR: firewall rule is still present." >&2
  exit 1
fi

rm -f "$STATE_FILE"

echo
echo "Demo DB access: CLOSED"
echo "Temporary firewall rules remaining: 0"