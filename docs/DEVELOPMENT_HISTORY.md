# Azure AI Governance Gateway — Development History

This document preserves implementation history, important deployment checkpoints and architecture decisions made during development.

It complements:

- [`../README.md`](../README.md) — current architecture and status;
- [`ROADMAP.md`](ROADMAP.md) — forward-looking plan.

The purpose is to avoid losing the reasoning behind changes as the demo evolves.

---

# Current checkpoint

Date:

```text
2026-09-05
```

Branch:

```text
stage-6-finops-guardrails
```

Latest Stage 6 application commits:

```text
706b0d5 Add monthly AI budget guardrails
e2cab50 Increase AI response write timeout
```

Current deployed Governance API image:

```text
acraigov0d60fe3d.azurecr.io/governance-api
@sha256:ccc3b32bfabd42e5f6daeb355152a8fc6e6f3ab63bfaf2a4dd6b363e587f3526
```

Current active Azure Container Apps revision:

```text
ca-governance-api-demo--0000010
```

Traffic:

```text
100%
```

Health:

```text
Healthy
```

Runtime:

```text
provider:   azure-openai
deployment: gpt-5-mini
```

Database migrations:

```text
1  initial_schema
2  usage_pricing_audit
3  budget_guardrails
```

Completed:

```text
Stage 6A -> FinOps accounting
Stage 6B -> Runtime budget guardrails
```

Next:

```text
Stage 6C -> Cost-aware Model Routing
```

---

# Azure environment

Subscription:

```text
Azure subscription 1
```

Subscription ID:

```text
74555f26-0688-4004-af97-8f82d11ae45a
```

Tenant:

```text
b586507e-cdea-4131-a68a-32a6bcd7634f
```

Region:

```text
swedencentral
```

Demo resource group:

```text
rg-aigov-demo
```

Monthly demo budget:

```text
€150
```

---

# Terraform state architecture

Backend resource group:

```text
rg-aigov-tfstate
```

Backend state layout:

```text
bootstrap.tfstate
demo.tfstate
demo-network-access.tfstate
demo-identity.tfstate
```

## `bootstrap.tfstate`

Owns Terraform backend/bootstrap infrastructure.

## `demo.tfstate`

Owns the main demo Azure infrastructure.

## `demo-network-access.tfstate`

Owns the PostgreSQL firewall rules required for Azure Container Apps outbound addresses.

At the time of isolation:

```text
161 firewall rules
```

were moved out of the main state.

## `demo-identity.tfstate`

Owns Microsoft Entra identity foundation resources.

---

# Major Azure resources

## Core platform

```text
Resource group:
  rg-aigov-demo

Log Analytics:
  log-aigov-demo

Application Insights:
  appi-aigov-demo

Key Vault:
  kv-aigov-0d60fe3d

Container Registry:
  acraigov0d60fe3d

Container Apps Environment:
  cae-aigov-demo
```

## PostgreSQL

```text
Server:
  psql-aigov-0d60fe3d

Version:
  PostgreSQL 16

SKU:
  B1ms

Storage:
  32 GB

Database:
  aigov

Administrator:
  aigovadmin

Backup retention:
  7 days
```

Demo network access uses explicit firewall rules.

A broad persistent:

```text
0.0.0.0
```

"Allow Azure services" rule is intentionally avoided.

## Governance API

```text
Container App:
  ca-governance-api-demo

User Assigned Managed Identity:
  id-governance-api-demo
```

Current Governance API identity roles include:

```text
AcrPull
Key Vault Secrets User
Cognitive Services OpenAI User
```

## Azure API Management

```text
APIM:
  apim-aigov-0d60fe3d

Tier:
  Consumption
```

APIM uses a System Assigned Managed Identity to authenticate to the backend.

## Azure OpenAI

```text
Account:
  aoai-aigov-0d60fe3d

Endpoint:
  https://aoai-aigov-0d60fe3d.openai.azure.com/

Deployment:
  gpt-5-mini

Model:
  gpt-5-mini

Version:
  2025-08-07

SKU:
  GlobalStandard

Capacity:
  10
```

Local authentication is disabled.

The Governance API uses Managed Identity.

---

# Stage 1 — Azure / Terraform foundation

The project began by establishing Terraform and cost-control foundations before introducing application runtime components.

Delivered:

- remote Terraform state;
- bootstrap resource group;
- storage account;
- state Blob container;
- demo resource group;
- Azure provider registration;
- €150 monthly Azure budget.

The project established the operating rule:

```text
terraform plan
      ->
review
      ->
terraform apply
      ->
verify
      ->
cost check
```

for new paid Azure resources.

---

# Stage 2 — Platform foundation

Created:

- Log Analytics;
- Application Insights;
- Azure Key Vault;
- Azure Container Registry;
- Azure Container Apps Environment.

The demo intentionally used low-cost SKUs and consumption-based services.

---

# Stage 3A — PostgreSQL

Provisioned Azure PostgreSQL Flexible Server 16.

Important cost choices:

```text
B1ms
32 GB
7-day backup
no HA
no geo-redundant backup
```

PostgreSQL credentials were stored in Key Vault.

---

# Stage 3B — Governance API

Created the Go Governance API.

Key runtime decisions:

- standard `net/http`;
- structured `slog` JSON logs;
- graceful shutdown;
- explicit HTTP timeouts;
- `pgxpool`;
- separate liveness/readiness;
- multi-stage Docker build;
- non-root Alpine runtime;
- immutable image deployment.

Endpoints introduced:

```text
GET /healthz
GET /readyz
```

---

# Stage 3C — Governance workflow

Introduced governance data model:

```text
governance_requests
policy_decisions
model_routes
usage_records
```

Introduced API:

```text
POST /v1/governance/requests
GET  /v1/governance/requests/{requestID}
```

Baseline policy:

```text
public       -> allow
internal     -> allow
confidential -> review
restricted   -> deny
```

Raw prompts were intentionally excluded from persistence.

Instead:

```text
SHA-256 hash
+
character count
```

are stored for audit purposes.

---

# Stage 3D — Network-access Terraform state isolation

## Problem

The Container Apps Consumption environment exposed a large list of possible outbound IPv4 addresses.

PostgreSQL firewall access required:

```text
161 rules
```

Terraform had to process those resources during normal demo plans.

## Decision

Move those rules from the main state to:

```text
demo-network-access.tfstate
```

## Outcome

The main demo state became substantially faster to plan.

This was an operational Terraform optimization without changing the Azure security model.

---

# Stage 4 — Entra ID and Azure API Management

The architecture moved from direct public API access toward the intended enterprise gateway topology.

## Entra API application

Governance API application/client ID:

```text
ecb0e3f3-ad8f-47d1-ae1d-5c8fb60ea1d1
```

Identifier URI:

```text
api://aigov-governance-api-demo
```

Scope:

```text
access_as_user
```

Scope UUID:

```text
a86d7a21-2cbc-4f23-9514-2c6d6e0ced31
```

## Demo client

Client ID:

```text
69bba583-8047-44c5-b1fb-bc8b36e4c3eb
```

Configured as a public client.

## APIM

APIM gateway:

```text
https://apim-aigov-0d60fe3d.azure-api.net
```

APIM identity principal:

```text
3d4ba4af-f91f-4db2-b085-2777f5b1bdfc
```

## Authentication design

APIM validates:

- Entra tenant;
- Governance API audience;
- demo client application;
- `access_as_user` scope.

Then APIM obtains its own backend token through Managed Identity.

Container Apps EasyAuth validates the APIM token.

## Verification

Verified behavior:

```text
direct anonymous backend call -> 401
direct valid user JWT         -> 403
APIM + valid JWT              -> 200
health/readiness              -> 200
```

This established two separate trust boundaries:

```text
user -> APIM
APIM -> backend
```

---

# Stage 5B — Governed AI provider gateway

Introduced a provider-neutral AI abstraction.

Initial provider:

```text
mock
```

Initial logical route:

```text
fast-general
```

New endpoint:

```text
POST /v1/ai/invoke
```

Governance semantics:

```text
allow  -> provider called
review -> provider NOT called
deny   -> provider NOT called
```

The mock provider returned deterministic content and synthetic token usage.

Routing was persisted in:

```text
model_routes
```

Usage was persisted in:

```text
usage_records
```

---

# Stage 5B.5 — Trust Entra caller identity from APIM

## Problem

A JSON `caller_subject` field is caller-controlled and therefore cannot be authoritative.

## Decision

APIM extracts:

```text
oid
```

from the validated JWT, with fallback to:

```text
sub
```

and overwrites:

```text
X-AIGOV-Caller-Subject
```

before forwarding.

The backend trusts this internal header over the JSON body.

## Result

A spoofing validation demonstrated that the persisted identity came from Entra rather than the caller-controlled payload.

---

# Stage 5C — Azure OpenAI

Stage 5C replaced mock-only execution with a real managed Azure OpenAI provider.

## Initial model attempt

The first intended model was:

```text
gpt-4o-mini
2024-07-18
```

Azure reported the deployment version as deprecated.

No failed deployment object was retained.

## Replacement

The model target changed to:

```text
gpt-5-mini
2025-08-07
GlobalStandard
```

This change required no enterprise-client API change because the client still targets the logical route:

```text
fast-general
```

## Managed Identity

Governance API UAMI:

```text
client ID:
52b5cb65-6791-4b92-8c61-bc08fcbc2516

principal ID:
2e8ce19c-34cf-4072-9921-13c0696e5c2d
```

Role:

```text
Cognitive Services OpenAI User
```

on the Azure OpenAI resource.

## Go provider

Added dependencies including:

```text
azidentity
openai-go/v3
```

Provider uses:

- explicit UAMI client ID;
- Azure Cognitive Services token scope;
- OpenAI-compatible Azure `/openai/v1/` base;
- Responses API;
- `store=false`;
- request timeout;
- retry policy.

## Authentication issue discovered

The first real invocation failed before the model request because the initial SDK wiring expected Azure endpoint middleware differently.

The provider was corrected to explicitly obtain Managed Identity tokens and attach Bearer authorization for the Azure OpenAI v1 endpoint.

## First successful real inference

Governed request ID:

```text
req_dadb142ad3fd6034b451f533178673df
```

Result:

```text
decision:        allow
logical route:   fast-general
provider:        azure-openai
model:           gpt-5-mini
input tokens:    28
output tokens:   152
estimated cost:  NULL
```

At that time FinOps pricing was intentionally not implemented.

The raw synthetic prompt marker was not found in logs or database persistence.

This completed Stage 5C.

---

# Stage 6A.1 — Provider-neutral FinOps pricing foundation

Branch:

```text
stage-6-finops-guardrails
```

Commit:

```text
a7ac52c Add provider-neutral FinOps pricing foundation
```

Created:

```text
internal/finops/pricing.go
internal/finops/calculator.go
```

with tests.

Initial design:

```text
Rate {
  Provider
  Model
  InputPerMillionUSD
  OutputPerMillionUSD
}
```

The catalog normalizes provider/model names, rejects invalid prices and detects duplicate rate entries.

The calculator returns an explicit known/unknown result rather than manufacturing zero for missing prices.

---

# Stage 6A.2 — Move cost estimation into FinOps

Commit:

```text
0577195 Move AI cost estimation into FinOps layer
```

## Design correction

Pricing was removed from:

```text
provider.Usage
```

The provider now reports usage only.

The router asks FinOps to calculate cost.

Result:

```text
known pricing   -> estimated_cost_usd
unknown pricing -> NULL
```

Mock pricing is an explicit real rate of zero and therefore remains a known cost of zero.

Azure pricing was initially omitted from the static catalog until the exact deployed SKU/rates were verified.

---

# Stage 6A.3 — Cached-token-aware Azure OpenAI pricing

Commit:

```text
f69ed83 Add cached-token-aware Azure OpenAI pricing
```

## Azure Retail Prices API discovery

For regular `GlobalStandard` `gpt-5-mini` in the deployed context:

```text
GPT 5 Mini Inpt Glbl
  $0.25 / 1M

GPT 5 Mini cchd Inpt Glbl
  $0.025 / 1M

GPT 5 Mini outpt Glbl
  $2.00 / 1M

effective:
  2025-08-01T00:00:00Z
```

Priority Processing rates were discovered separately but intentionally not used.

## Rate model extended

Added:

```text
CachedInputPerMillionUSD
Source
EffectiveStartDate
```

## Usage model extended

Added:

```text
CachedInputTokens
```

## Azure provider mapping

Azure OpenAI response usage now maps:

```text
response.Usage.InputTokensDetails.CachedTokens
```

## Calculation

```text
non_cached = input - cached

cost =
  non_cached * regular_input_rate
  + cached * cached_input_rate
  + output * output_rate
```

Validation:

```text
0 <= cached <= input
```

---

# Stage 6A.4 — Pricing audit persistence

Commit:

```text
97d11e8 Persist FinOps pricing audit metadata
```

## API audit model

Added a pricing snapshot containing:

```text
source
effective_start_date
input_per_million_usd
cached_input_per_million_usd
output_per_million_usd
```

The snapshot exists only when pricing is known.

## Database migration

Added:

```text
000002_usage_pricing_audit.up.sql
000002_usage_pricing_audit.down.sql
```

New `usage_records` fields:

```text
cached_input_tokens
pricing_source
pricing_effective_start_date
input_price_per_million_usd
cached_input_price_per_million_usd
output_price_per_million_usd
```

Added validation constraints for:

- non-negative cached tokens;
- cached tokens not exceeding input tokens;
- non-negative pricing rates.

## Historical integrity

No default values were introduced for the new columns.

This was deliberate.

Historical records from before this migration remain:

```text
NULL
```

where cached-token/pricing information did not exist.

## Local migration verification

The migration was tested:

```text
UP
verify
DOWN
verify removed
UP
verify restored
```

A post-reapply mock inference correctly persisted:

```text
cached_input_tokens = 0
estimated_cost_usd  = 0
pricing_source      = local mock
rates               = 0
```

---

# Generic migration runner

Stage 6A.4 replaced hardcoded Azure migration behavior with:

```text
scripts/run-migrations.sh
```

The runner:

- discovers `*.up.sql`;
- validates filenames;
- sorts by migration version;
- reads `schema_migrations`;
- skips already applied migrations;
- detects migration-name mismatches;
- applies new migrations transactionally;
- records version/name.

## Safety guard

During local testing an important scenario was discovered:

```text
schema exists
but
schema_migrations does not exist
```

The generic runner was therefore hardened to stop in this state rather than automatically reapplying migration `000001`.

Local baseline was registered explicitly only after verifying all expected core tables existed.

## Local runner validation

Verified:

```text
000001 -> already applied
000002 -> applied
second run -> both skipped
```

This proved idempotency.

---

# Azure migration `000002`

Before applying the migration, a read-only preflight verified:

```text
schema_migrations exists: t
core table count:          4
migration 000001:          initial_schema
migration 000002:          not applied
pricing audit columns:     0/6
```

Result:

```text
SAFE TO APPLY 000002
```

The Azure migration runner then applied:

```text
000002_usage_pricing_audit
```

and recorded:

```text
1 initial_schema
2 usage_pricing_audit
```

A second run skipped both migrations, proving Azure idempotency.

Azure verification confirmed:

```text
pricing columns:        6/6
pricing constraints:    5/5
historical fields NULL: true
```

Temporary PostgreSQL firewall rules were removed after the session.

---

# Stage 6A.5 — FinOps-enabled Azure deployment

The FinOps-enabled Governance API was deployed and then verified with real Azure OpenAI usage.

Initial Stage 6A deployment used an immutable container image and an in-place Container App update with:

```text
0 add
1 change
0 destroy
```

The only Terraform semantic change was the image digest.

## Final Stage 6A verification

A fresh synthetic Azure OpenAI request confirmed:

```text
caller derived from Entra oid
governance decision = allow
provider_called = true
provider = azure-openai
model = gpt-5-mini
pricing source = Azure Retail Prices API
pricing effective date = 2025-08-01
```

Provider-authoritative token usage and the FinOps cost estimate were persisted in PostgreSQL.

Historical unknown pricing fields remained `NULL`.

## Pricing-audit persistence correction

A pointer-aliasing persistence defect caused one synthetic row to retain the correct estimated cost but incorrect repeated rate values in the pricing snapshot.

The defect was fixed in:

```text
514ff24 Fix FinOps pricing audit persistence
```

A fresh request verified correct input, cached-input and output rate persistence.

The historical synthetic row was not broadly backfilled because the project preserves historical audit truth rather than rewriting records with inferred data.

This completed Stage 6A.

---

# Stage 6B — Runtime budget guardrails

Stage 6B introduced monthly cost-center financial governance before AI provider invocation.

Application/database commit:

```text
706b0d5 Add monthly AI budget guardrails
```

## Data model

Migration:

```text
000003_budget_guardrails
```

introduced:

```text
budget_policies
budget_decisions
```

`budget_policies` defines one enabled policy per cost center, including:

```text
currency = USD
monthly_limit_usd
review_threshold_percent
```

`budget_decisions` snapshots the policy and financial state used for a governance-allowed request.

## Spend calculation

The budget repository calculates current monthly spend from persisted `usage_records`, joined through `governance_requests.cost_center`.

The period is the UTC calendar month.

Known costs are summed separately from unknown-cost rows.

Unknown cost is never treated as zero.

## Decision behavior

```text
missing cost center       -> review
missing active policy     -> review
unknown-cost usage exists -> review
monthly limit = 0         -> deny
spent >= monthly limit    -> deny
utilization >= threshold  -> review
otherwise                 -> allow
```

Budget evaluation executes after governance `allow` and before model routing/provider invocation.

Governance `review` or `deny` stops before the budget service.

Budget `review` or `deny` stops before route/provider/usage persistence.

## Local verification

Local HTTP validation completed:

```text
allow  -> HTTP 200
review -> HTTP 202
deny   -> HTTP 403
```

Database invariants confirmed:

```text
ALLOW  -> budget decision + route + usage
REVIEW -> budget decision only
DENY   -> budget decision only
```

## Azure migration and policy seed

Migration `000003` was applied to Azure PostgreSQL and recorded in `schema_migrations`.

The migration timestamp was:

```text
2026-09-05 19:50:13.631984+00
```

Synthetic policies were seeded for:

```text
BUDGET-ALLOW
  limit = 100 USD
  review threshold = 80%

BUDGET-REVIEW
  limit = 100 USD
  review threshold = 0%

BUDGET-DENY
  limit = 0 USD
  review threshold = 80%
```

All direct PostgreSQL sessions used a temporary firewall rule restricted to the exact current public IPv4 address.

Cleanup verification confirmed no Stage 6B or migration temporary firewall rules remained.

## First Azure E2E attempt and timeout diagnosis

The first Azure `ALLOW` E2E request returned an empty HTTP `500`.

Backend logs nevertheless showed:

```text
governance_decision = allow
budget_decision = allow
effective_decision = allow
provider_called = true
```

with no backend error.

The Go server was configured with:

```text
WriteTimeout = 15 seconds
```

A short synthetic allow probe completed successfully in approximately 2.34 seconds and returned HTTP `200`.

The failed request's exact duration was not measured, so the timeout diagnosis was not treated as absolute proof. However, the combination of an empty downstream response, successful backend completion, no application error, the 15-second write timeout and successful shorter probes made the write timeout the leading diagnosis.

The server timeout was changed to:

```text
WriteTimeout = 120 seconds
```

in:

```text
e2cab50 Increase AI response write timeout
```

## Final Stage 6B Azure deployment

Timeout-hotfix image:

```text
acraigov0d60fe3d.azurecr.io/governance-api
@sha256:ccc3b32bfabd42e5f6daeb355152a8fc6e6f3ab63bfaf2a4dd6b363e587f3526
```

Terraform plan:

```text
0 to add
1 to change
0 to destroy
```

Only:

```text
azurerm_container_app.governance_api
```

was updated in-place.

Final revision:

```text
ca-governance-api-demo--0000010
```

Status:

```text
Healthy
100% traffic
```

Health and readiness both returned HTTP `200`.

A post-apply Terraform plan returned:

```text
No changes.
exit code = 0
```

## Final APIM E2E validation

Final successful `ALLOW`:

```text
request_id = req_59dc5a9d8abecb10f132a032bc32a9c0
HTTP = 200
provider_called = true
provider = azure-openai
model = gpt-5-mini
input_tokens = 15
cached_input_tokens = 0
output_tokens = 97
estimated_cost_usd = 0.00019775
```

Final `REVIEW`:

```text
request_id = req_e60bd92421c0ff01de17daba41a02e43
HTTP = 202
provider_called = false
```

Final `DENY`:

```text
request_id = req_5e5f9778a954e859aeb29f1b91fef504
HTTP = 403
provider_called = false
```

## Final PostgreSQL audit

The database audit confirmed:

```text
ALLOW  -> 1 budget decision, 1 model route, 1 usage record
REVIEW -> 1 budget decision, 0 model routes, 0 usage records
DENY   -> 1 budget decision, 0 model routes, 0 usage records
```

The final allow record persisted:

```text
provider = azure-openai
model = gpt-5-mini
input tokens = 15
cached input tokens = 0
output tokens = 97
estimated cost = 0.00019775 USD
pricing source = Azure Retail Prices API
effective date = 2025-08-01
input rate = 0.25 USD / 1M
cached input rate = 0.025 USD / 1M
output rate = 2.00 USD / 1M
```

Raw prompt-content checks returned zero matches across:

```text
governance_requests
policy_decisions
budget_decisions
model_routes
usage_records
```

## Budget concurrency limitation

The Stage 6B MVP evaluates accrued usage cost but does not reserve the expected cost of a request before provider invocation.

Therefore the monthly limit is not an atomic hard cap.

The current request or concurrent requests may temporarily move spend above the configured threshold before a subsequent request observes the updated accrued balance.

A production hard-cap design should introduce a reservation/ledger or another concurrency-safe financial pre-authorization mechanism.

---

# Security decisions preserved

## No raw prompt persistence

Only prompt hash and length are stored.

The Stage 6B Azure database audit again confirmed that synthetic raw prompt markers were absent from governance, policy, budget, routing and usage records.

## No Azure OpenAI API keys

Managed Identity is used.

## APIM is the enterprise entry point

End-user authentication is validated by APIM.

## Backend has a separate identity boundary

Container Apps EasyAuth allows the APIM workload identity.

## Caller subject is token-derived

The backend does not trust a caller-provided JSON identity.

## Governance and budget stop before provider invocation

Governance `review` / `deny` stops before budget and provider execution.

Budget `review` / `deny` stops before model routing and provider execution.

## Database public access is constrained

The demo uses explicit IP rules.

Migration, seed, validation and DBeaver access use temporary exact-IP rules.

No persistent broad `0.0.0.0` access is intentionally introduced.

## Unknown pricing is not zero

Missing pricing remains:

```text
NULL
```

For Stage 6B, any unknown-cost usage record in the active budget period forces `review`.

## Missing financial policy fails closed

Missing cost center or missing enabled budget policy results in `review`, not implicit permission.

## Historical audit evidence is not rewritten

New fields remain `NULL` for historical records where the information was unavailable.

---

# Cost and production trade-offs

The environment is intentionally a demo.

Low-cost choices include:

```text
APIM Consumption
Container Apps Consumption
PostgreSQL B1ms
ACR Basic
Azure OpenAI pay-per-token
```

These choices are not necessarily production recommendations.

## Production areas still requiring hardening

- private networking;
- controlled egress;
- HA / DR;
- production database sizing;
- centralized policy configuration;
- SIEM/security monitoring;
- CI/CD gates;
- provider/model lifecycle governance;
- data-residency-aware deployment choices;
- enterprise billing integration.

---

# Selected Git history

## Stage 3D

```text
203c5d6 Isolate demo network access Terraform state
```

Merged through PR #2.

## Stage 4

```text
c25aeb0 Add Microsoft Entra identity foundation
49fc1f3 Add APIM Consumption foundation
967370b Publish governance API through APIM
56769d0 Protect governance API with Entra authentication
d064992 Require APIM identity for governance backend
```

Merged through PR #3.

## Stage 5B

```text
a69e239 Add governed AI invocation with mock provider
b5ac793 Deploy governed AI invocation through APIM
```

Merged through PR #4.

## Stage 5B.5

```text
1a3935e Trust Entra caller identity from APIM
```

Merged through PR #5.

## Stage 5C

```text
55231cc Provision Azure OpenAI with managed identity
e4a5018 Add Azure OpenAI managed identity provider
4fbb36b Fix Azure OpenAI v1 managed identity auth
556d94e Deploy Azure OpenAI provider runtime
```

Merged through PR #6.

## Stage 6A

```text
a7ac52c Add provider-neutral FinOps pricing foundation
0577195 Move AI cost estimation into FinOps layer
f69ed83 Add cached-token-aware Azure OpenAI pricing
97d11e8 Persist FinOps pricing audit metadata
514ff24 Fix FinOps pricing audit persistence
```

## Stage 6B

```text
706b0d5 Add monthly AI budget guardrails
e2cab50 Increase AI response write timeout
```

Current Stage 6 branch:

```text
stage-6-finops-guardrails
```

---

# Next checkpoint

Stage 6A FinOps accounting and Stage 6B runtime budget guardrails are complete and verified in Azure.

Next:

```text
Stage 6C -> Cost-aware Model Routing
```

Stage 6C should combine:

```text
governance policy
+
data classification
+
model/provider capability
+
financial policy
+
provider/model economics
```

while preserving explainable routing decisions and immutable financial/routing audit evidence.

The Stage 6B accrued-spend concurrency limitation remains explicit future hardening work for any production-grade atomic budget-cap implementation.
