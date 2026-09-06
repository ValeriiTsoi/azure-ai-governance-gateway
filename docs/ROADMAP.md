# Azure AI Governance Gateway — Development Roadmap

This roadmap tracks the incremental delivery of the Azure AI Governance Gateway reference implementation.

The project follows a simple delivery rule:

```text
plan -> implement -> test -> review -> deploy -> verify -> cost check
```

The roadmap is intentionally pragmatic. Stages may be reordered when a real-client proof is more valuable than adding another internal capability.

---

## Current status

| Stage | Scope | Status |
|---|---|:---:|
| 1 | Azure / Terraform foundation | ✅ Complete |
| 2 | Managed platform foundation | ✅ Complete |
| 3A | PostgreSQL foundation | ✅ Complete |
| 3B | Governance API runtime | ✅ Complete |
| 3C | Governance workflow | ✅ Complete |
| 3D | Terraform network-access state isolation | ✅ Complete |
| 4 | Microsoft Entra ID + Azure API Management | ✅ Complete |
| 5B | Provider-neutral governed AI invocation | ✅ Complete |
| 5B.5 | Trusted caller identity propagation | ✅ Complete |
| 5C | Azure OpenAI Managed Identity provider | ✅ Complete |
| 6A | Provider-neutral FinOps accounting | ✅ Complete |
| 6B | Runtime budget guardrails | ✅ Complete |
| 6C | Cost-aware multi-model routing | ⬜ Planned |
| 7A | OpenAI-compatible facade | ✅ Complete |
| 7B | IDE/API gateway credential + dual auth | ✅ Complete |
| 7C | Cursor client integration | ⚠️ Client-plan limitation |
| 7D | VS Code + Continue Chat | ✅ Complete |
| 7E | Dedicated agent identities + tools/function calling | ⬜ Planned |

Stage 7 was intentionally pulled forward before Stage 6C to prove the key platform assumption: **real developer tools and API clients can use the same governed execution path without bypassing governance, budget, routing, FinOps or audit controls.**

---

# Stage 1 — Azure / Terraform foundation

## Objective

Establish a reproducible Azure foundation before application development.

## Delivered

- Azure provider configuration;
- Terraform remote state;
- state resource group and storage account;
- demo resource group;
- required Azure provider registration;
- monthly Azure Cost Management budget.

## Operating rule

New paid Azure resources follow:

```text
terraform plan
  -> review
  -> terraform apply
  -> verify
  -> cost check
```

---

# Stage 2 — Managed platform foundation

## Objective

Create the minimum Azure services needed to host and observe the governance control plane.

## Delivered

- Log Analytics;
- Application Insights;
- Azure Key Vault;
- Azure Container Registry Basic;
- Azure Container Apps Environment;
- Consumption workload profile.

The demo prefers consumption-oriented managed services where they preserve the architecture and reduce operating cost.

---

# Stage 3 — Governance runtime and data foundation

## Stage 3A — PostgreSQL

Delivered:

- Azure Database for PostgreSQL Flexible Server 16;
- demo database;
- Key Vault-backed administrator credential;
- explicit firewall rules rather than broad `0.0.0.0` access;
- low-cost demo sizing.

Production divergence: use private networking and controlled egress rather than the current public-endpoint demo pattern.

## Stage 3B — Governance API runtime

Delivered:

- Go service;
- structured logging;
- graceful shutdown and HTTP timeouts;
- PostgreSQL connection pool;
- `/healthz` and `/readyz`;
- Docker multi-stage build;
- non-root runtime;
- Azure Container Apps deployment;
- User Assigned Managed Identity;
- immutable OCI digest deployment.

## Stage 3C — Governance workflow

Delivered:

```text
POST /v1/governance/requests
GET  /v1/governance/requests/{requestID}
```

Initial decisions:

```text
allow
review
deny
```

Initial classifications:

```text
public
internal
confidential
restricted
```

Baseline behavior:

```text
public       -> allow
internal     -> allow
confidential -> review
restricted   -> deny
```

Audit behavior deliberately avoids raw-prompt persistence.

## Stage 3D — Network-access state isolation

Azure Container Apps Consumption exposed a large outbound-IP set. Keeping every PostgreSQL firewall rule in the main Terraform state made normal plans unnecessarily slow.

The firewall allowlist was therefore isolated into:

```text
demo-network-access.tfstate
```

Current state layout:

```text
bootstrap.tfstate
demo.tfstate
demo-network-access.tfstate
demo-identity.tfstate
```

This is a demo optimization, not a target production networking topology.

---

# Stage 4 — Microsoft Entra ID + Azure API Management

## Objective

Move from direct application access to an enterprise gateway and identity boundary.

## Delivered

### Microsoft Entra ID

- Governance API application registration;
- delegated `access_as_user` scope;
- public demo client for interactive/device-code testing.

### Azure API Management

APIM responsibilities:

- validate Microsoft Entra tokens;
- validate tenant, audience, client and delegated scope;
- derive a trusted caller subject from `oid` with `sub` fallback;
- overwrite internal caller headers;
- publish governance and governed AI APIs.

### Backend identity boundary

APIM authenticates to the Governance API using its Managed Identity.

Container Apps EasyAuth protects the backend so the external caller token is not itself sufficient for direct backend access.

---

# Stage 5 — Governed AI provider execution

## Stage 5B — Provider-neutral invocation

Delivered:

- provider-neutral Go interface;
- deterministic mock provider;
- logical route `fast-general`;
- `POST /v1/ai/invoke`;
- governance before provider execution;
- model-route persistence;
- usage persistence.

## Stage 5B.5 — Trusted caller identity

The original API allowed caller identity to be supplied in client JSON.

That was changed so APIM derives and overwrites the trusted caller identity after token validation. The backend therefore does not trust a user-controlled caller field.

## Stage 5C — Azure OpenAI provider

Delivered:

- Azure OpenAI account and deployment;
- `gpt-5-mini` route;
- local model-key authentication disabled;
- Governance API Managed Identity with Azure OpenAI RBAC;
- real Azure inference;
- provider-authoritative token usage;
- cached input token capture where reported.

The logical client route remained stable while the concrete provider changed from mock to Azure OpenAI.

---

# Stage 6 — FinOps and governance guardrails

## Stage 6A — Provider-neutral FinOps accounting

### Architecture decision

Pricing is not owned by provider adapters.

```text
provider -> reports usage
finops   -> owns prices and cost calculation
router   -> orchestrates
```

This prevents provider implementation details from becoming the source of mutable financial policy.

### Delivered

- provider-neutral pricing catalog;
- token-based cost calculator;
- cached-input-aware cost calculation;
- Azure Retail Prices API rate discovery;
- pricing source/effective-date metadata;
- immutable pricing snapshots in usage audit records;
- generic numbered migration runner;
- migration `000002_usage_pricing_audit`;
- historical rows preserve unavailable pricing fields as `NULL`.

### Core invariant

Unknown pricing is not zero.

```text
unknown price -> NULL
```

## Stage 6B — Runtime budget guardrails

### Objective

Evaluate monthly cost-center spend before model routing and provider invocation.

### Delivered

- versioned `budget_policies`;
- immutable `budget_decisions`;
- monthly UTC periods;
- accrued-spend calculation;
- `allow`, `review`, `deny` outcomes;
- fail-closed behavior for missing financial policy;
- unknown-cost usage forces review;
- review/deny stop before model routing and provider execution;
- local and Azure end-to-end verification;
- migration `000003_budget_guardrails`.

### Execution semantics

```text
governance allow
  -> budget allow
  -> route
  -> provider
```

```text
governance review/deny
  -> stop
```

```text
budget review/deny
  -> stop before route/provider
```

### Current limitation

The MVP evaluates accrued spend but does not reserve the expected cost of in-flight requests. It is therefore not an atomic hard cap under concurrency.

Future production hardening should add a reservation/ledger or another concurrency-safe financial pre-authorization mechanism.

---

# Stage 6C — Cost-aware multi-model routing

## Status

Planned; intentionally deferred while Stage 7 proves real-client integration.

## Target inputs

```text
governance policy
+ data classification
+ model/provider capability
+ budget state
+ pricing/economics
+ availability/health
```

## Target behavior

The client should continue to request a logical capability while the gateway chooses an approved concrete model/provider and records an explainable routing reason.

Examples:

- prefer the lowest-cost approved route that satisfies the requested capability;
- avoid a route that violates data classification policy;
- optionally downgrade within policy when budget pressure increases;
- preserve model/provider economics and routing rationale in immutable audit evidence.

---

# Stage 7 — Real client integration

## Why Stage 7 was pulled forward

The internal governance path was already proven through API tests. The next highest-value question was whether real developer tools could consume the gateway without receiving direct provider credentials or bypassing governance controls.

The target architecture is:

```text
IDE / SDK / agent client
 -> OpenAI-compatible API
 -> APIM
 -> trusted client identity
 -> Governance API
 -> policy
 -> budget
 -> route
 -> Azure OpenAI
 -> usage + FinOps + audit
```

## Stage 7A — OpenAI-compatible facade

### Delivered

```text
GET  /openai/v1/models
POST /openai/v1/chat/completions
```

External model alias:

```text
aigov-fast-general
```

Internal logical route:

```text
fast-general
```

Concrete provider route:

```text
azure-openai / gpt-5-mini
```

The facade converts standard `messages[]` into the existing governed invocation request and returns an OpenAI-like `chat.completion` response.

### Current compatibility boundary

Stage 7A intentionally supports:

- Chat Completions;
- standard `messages[]` roles used by chat clients;
- `stream=false`;
- no native `tools`/function-calling yet.

The adapter does not create a second governance pipeline.

### Verification

Azure E2E verified:

- delegated Entra token -> APIM -> backend -> Azure OpenAI;
- logical model alias -> `gpt-5-mini` route;
- provider-authoritative usage;
- FinOps pricing/cost persistence;
- raw prompt marker absent from PostgreSQL.

## Stage 7B — IDE/API gateway credential and dual auth

### Objective

Support clients that can send an OpenAI-style API key but cannot perform Microsoft Entra interactive/device-code OAuth themselves.

### Delivered

- separate demo gateway credential stored in Azure Key Vault;
- APIM Managed Identity granted `Key Vault Secrets User`;
- APIM Named Value references the Key Vault secret by versionless URI;
- OpenAI-compatible APIM policy supports two auth modes:
  - gateway demo credential;
  - existing delegated Microsoft Entra token;
- APIM maps the demo credential to trusted governance headers;
- Azure OpenAI credentials remain hidden from clients.

The secret value is not stored in Terraform source.

## Stage 7C — Cursor

Cursor was configured successfully with:

- custom OpenAI base URL;
- gateway API key;
- custom model alias.

The free client plan used for the demo did not permit selecting the custom model in Chat/Agent, so Cursor is not required for the public demo.

This is a **client-plan limitation**, not a gateway compatibility failure.

## Stage 7D — VS Code + Continue

### Delivered

VS Code + Continue was configured as a real IDE client using:

```yaml
provider: openai
model: aigov-fast-general
apiBase: https://<apim>.azure-api.net/openai/v1
useResponsesApi: false
```

The gateway credential is loaded from Continue local secrets rather than committed to the project.

Interactive IDE Chat successfully reached the governed Azure path and returned model-generated code/content.

### Demo use

The public demo uses Continue Chat to show a normal developer experience and then queries PostgreSQL audit evidence to demonstrate:

- trusted caller/application context;
- governance decision;
- budget decision;
- routed model/provider;
- token usage;
- estimated cost;
- pricing provenance;
- absence of raw prompt content.

## Stage 7E — Agent clients and richer OpenAI compatibility

### Planned

- dedicated per-client identities/cost centers such as `continue-demo` and `agent-demo`;
- Python/OpenAI SDK example;
- explicit agent-client demo;
- native streaming (`stream=true` / SSE);
- tools/function-calling compatibility;
- optional Responses API facade where useful;
- richer model metadata;
- explicit application-to-governance-profile mapping;
- budget allow/review/deny demonstration by client identity.

---

# Capability backlog

## Policy administration

- move baseline policy data from code into governed configuration;
- version policy changes;
- admin workflow and approvals;
- policy simulation before activation.

## Governance dashboards

- requests by caller/application;
- policy allow/review/deny distribution;
- budget decisions;
- usage and cost by cost center;
- model/provider routing;
- pricing source and effective dates.

## CI/CD

Initial public-repository CI can validate source quality without cloud credentials:

- Go tests;
- Go vet;
- Terraform formatting;
- shell syntax checks.

Future delivery work should add controlled deployment promotion and security gates.

## Production networking

- private endpoints where appropriate;
- VNet integration;
- controlled/stable egress;
- private PostgreSQL connectivity;
- production APIM topology.

## Resilience

- multiple provider/model routes;
- health-aware routing;
- retry/circuit-breaker strategy;
- regional failover design;
- database HA/DR.

## Enterprise security

- dedicated application identities;
- gateway credential rotation;
- private vulnerability reporting;
- SIEM integration;
- security event correlation;
- rate limiting / abuse controls;
- model safety controls.

---

# Roadmap principles

## 1. Governance before invocation

A model should not be called until governance and financial controls permit it.

## 2. Provider neutrality

Clients should depend on logical platform capabilities rather than concrete model vendors.

## 3. Pricing belongs outside providers

Provider adapters report usage; the FinOps layer owns mutable pricing and cost calculation.

## 4. Unknown is not zero

Missing financial evidence remains unknown and must not be fabricated.

## 5. Preserve historical truth

New schema fields remain `NULL` for historical rows when the information did not exist at the time.

## 6. Immutable deployment artifacts

Azure runtime deployments use immutable container image digests.

## 7. Review paid infrastructure changes

Terraform plan review precedes apply for paid Azure resources.

## 8. Demo architecture is not automatically production architecture

Low-cost public networking, small SKUs and single-region choices are explicitly documented as demo trade-offs.

## 9. Raw prompt minimization

Audit should retain governance evidence without unnecessarily persisting prompt content.

## 10. Real clients must reuse the same governance path

OpenAI compatibility is an adapter at the edge, not a second bypass pipeline.
