#!/usr/bin/env bash

set -euo pipefail

PROVIDERS=(
  "Microsoft.Storage"
  "Microsoft.Consumption"
  "Microsoft.Insights"
  "Microsoft.OperationalInsights"
  "Microsoft.KeyVault"
  "Microsoft.App"
  "Microsoft.ContainerRegistry"
  "Microsoft.EventHub"
  "Microsoft.DBforPostgreSQL"
  "Microsoft.Search"
  "Microsoft.CognitiveServices"
  "Microsoft.ApiManagement"
)

echo "Azure subscription:"
az account show \
  --query "{Name:name,State:state}" \
  --output table

echo

for provider in "${PROVIDERS[@]}"; do
  state="$(az provider show \
    --namespace "$provider" \
    --query registrationState \
    --output tsv 2>/dev/null || true)"

  if [[ "$state" == "Registered" ]]; then
    echo "[OK]       $provider"
  else
    echo "[REGISTER] $provider"
    az provider register \
      --namespace "$provider" \
      --wait
  fi
done

echo
echo "Provider registration complete."
