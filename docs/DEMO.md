# Public Demo Runbook

This runbook demonstrates the Azure AI Governance Gateway in approximately 2–3 minutes.

The goal is to show a normal developer experience first, then prove that the request was governed, routed, metered and audited without persisting the raw prompt.

> Use synthetic, non-sensitive prompts only.

---

## Demo story

```text
VS Code + Continue
        |
        v
OpenAI-compatible APIM endpoint
        |
        v
Governance API
        |
        +--> policy decision
        +--> cost-center budget decision
        +--> logical model routing
        |
        v
Azure OpenAI
        |
        v
provider-authoritative usage + FinOps cost + PostgreSQL audit
```

The same endpoint can also be consumed by API/agent-facing clients.

---

## Before the audience joins

Open temporary database access from the exact current public IPv4 address:

```bash
./scripts/demo-db-access-open.sh
```

Wait for:

```text
Demo DB access: READY
```

Do not leave this rule open after the demo.

Make sure VS Code is open on this repository and Continue is configured with the governed model:

```text
AIGOV Fast General
```

---

# 1. VS Code + Continue

In Continue Chat, send:

```text
DEMO_VSCODE_001. Inspect this repository and explain in 3 short bullets how an AI request is governed before it reaches Azure OpenAI. Do not repeat the marker and do not modify files.
```

Suggested narration:

> The developer works normally in VS Code. Continue calls an OpenAI-compatible endpoint, but the request goes to our enterprise gateway rather than directly to a model provider.

The request path is:

```text
Continue
 -> APIM
 -> Governance API
 -> policy
 -> budget
 -> route
 -> Azure OpenAI gpt-5-mini
```

---

# 2. API / agent-facing client

Load the demo gateway credential from Key Vault without printing it:

```bash
export AIGOV_DEMO_KEY="$(
  az keyvault secret show \
    --vault-name "${KEY_VAULT_NAME:-kv-aigov-0d60fe3d}" \
    --name "${AIGOV_DEMO_KEY_SECRET_NAME:-cursor-demo-api-key}" \
    --query value \
    -o tsv
)"
```

Set the OpenAI-compatible base URL if it is not already exported:

```bash
export AIGOV_OPENAI_BASE_URL="${AIGOV_OPENAI_BASE_URL:-https://apim-aigov-0d60fe3d.azure-api.net/openai/v1}"
```

Call the same governed interface:

```bash
curl -sS \
  "$AIGOV_OPENAI_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $AIGOV_DEMO_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "aigov-fast-general",
    "messages": [
      {
        "role": "system",
        "content": "You are a concise enterprise software engineering assistant."
      },
      {
        "role": "user",
        "content": "DEMO_AGENT_001. Give a concise 3-step plan for adding a new governed data classification named restricted-partner. Do not repeat the marker."
      }
    ],
    "stream": false
  }' \
| jq '{
    id,
    model,
    answer: .choices[0].message.content,
    usage
  }'
```

Then remove the gateway credential from the shell:

```bash
unset AIGOV_DEMO_KEY
```

Suggested narration:

> The same governed endpoint is available to API and agent clients through a standard OpenAI-compatible contract. The client never receives an Azure OpenAI key.

> Current Stage 7 supports non-streaming Chat Completions. Native tools/function calling is planned follow-up work.

---

# 3. Show governance + FinOps audit

Load the PostgreSQL password from Key Vault without printing it:

```bash
export PGPASSWORD="$(
  az keyvault secret show \
    --vault-name "${KEY_VAULT_NAME:-kv-aigov-0d60fe3d}" \
    --name "${POSTGRES_PASSWORD_SECRET_NAME:-postgresql-admin-password}" \
    --query value \
    -o tsv
)"
```

Show the two most recent OpenAI-compatible requests:

```bash
docker run --rm \
  -e PGPASSWORD \
  postgres:16-alpine \
  psql \
    "host=${POSTGRES_SERVER_NAME:-psql-aigov-0d60fe3d}.postgres.database.azure.com port=5432 dbname=${POSTGRES_DATABASE_NAME:-aigov} user=${POSTGRES_ADMIN_LOGIN:-aigovadmin} sslmode=require" \
    -P pager=off \
    -c "
SELECT
    gr.created_at,
    gr.request_id,
    gr.caller_subject,
    gr.cost_center,
    gr.requested_model,
    pd.decision AS governance,
    bd.decision AS budget,
    mr.routed_model,
    mr.provider,
    ur.input_tokens,
    ur.output_tokens,
    ur.estimated_cost_usd,
    ur.pricing_source
FROM governance_requests gr
LEFT JOIN policy_decisions pd
    ON pd.governance_request_id = gr.id
LEFT JOIN budget_decisions bd
    ON bd.governance_request_id = gr.id
LEFT JOIN model_routes mr
    ON mr.governance_request_id = gr.id
LEFT JOIN usage_records ur
    ON ur.governance_request_id = gr.id
WHERE gr.use_case = 'openai-chat-completions'
ORDER BY gr.created_at DESC
LIMIT 2;
"
```

Highlight:

```text
governance   = allow
budget       = allow
routed_model = gpt-5-mini
provider     = azure-openai
usage        = provider-authoritative tokens
cost         = non-zero when pricing is known
pricing      = Azure Retail Prices API
```

Suggested narration:

> Governance and financial controls run before provider invocation. After an allowed request completes, the platform records provider-authoritative usage and the pricing evidence used for cost attribution.

---

# 4. Prove raw prompts are not persisted

Search the database dump for the synthetic markers:

```bash
MATCHES="$(
  docker run --rm \
    -e PGPASSWORD \
    postgres:16-alpine \
    pg_dump \
      "host=${POSTGRES_SERVER_NAME:-psql-aigov-0d60fe3d}.postgres.database.azure.com port=5432 dbname=${POSTGRES_DATABASE_NAME:-aigov} user=${POSTGRES_ADMIN_LOGIN:-aigovadmin} sslmode=require" \
      --data-only \
  | grep -Ec 'DEMO_VSCODE_001|DEMO_AGENT_001' \
  || true
)"

echo "raw prompt markers stored in DB: $MATCHES"
```

Expected:

```text
raw prompt markers stored in DB: 0
```

Suggested narration:

> We keep governance, routing, usage and cost metadata, but the Governance API does not persist the raw prompt.

---

# 5. Cleanup

Remove the password from the shell:

```bash
unset PGPASSWORD
```

Close temporary database access:

```bash
./scripts/demo-db-access-close.sh
```

Expected:

```text
Demo DB access: CLOSED
Temporary firewall rules remaining: 0
```

---

## What this demo proves

The live path demonstrates:

```text
real IDE/API client
+ centralized authentication
+ governance decision
+ budget decision
+ provider-neutral route
+ real Azure OpenAI execution
+ authoritative usage
+ FinOps cost attribution
+ immutable audit metadata
+ raw-prompt minimization
```

It does **not** claim that the current demo topology is production hardened. Private networking, HA/DR, native streaming, tool/function calling, per-client demo identities and stricter financial reservation semantics remain roadmap items.
