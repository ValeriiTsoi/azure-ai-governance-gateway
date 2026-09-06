# Security Policy

Azure AI Governance Gateway is a reference architecture and demonstration project.

## Reporting a vulnerability

Please do **not** open a public GitHub issue containing credentials, tokens, private infrastructure details or a reproducible security exploit.

Use GitHub private vulnerability reporting if it is enabled for this repository. If private reporting is unavailable, contact the repository maintainer privately through the GitHub profile before sharing sensitive details.

## Secret handling

The repository is designed around the following rules:

- do not commit Terraform state or plan files;
- do not commit `.env` files;
- do not commit Azure access tokens, API keys, database passwords or connection strings;
- do not commit local Continue/Cursor/IDE configuration containing credentials;
- use Azure Key Vault for runtime/demo secrets;
- use Managed Identity and Azure RBAC where supported;
- do not expose Azure OpenAI keys to client applications;
- rotate/revoke any credential that may have been exposed before attempting Git-history cleanup.

Terraform state deserves special attention: even when Terraform configuration contains no secret literals, state and plan files can contain provider-generated credentials, connection strings and sensitive resource attributes.

## Demo database access

The demo PostgreSQL endpoint uses explicit firewall rules.

Temporary local access must:

1. determine the current public IPv4 address;
2. create a rule where start IP equals end IP;
3. avoid broad `0.0.0.0` / "Allow Azure services" access;
4. remove the temporary rule immediately after the operation;
5. verify that the rule no longer exists.

The helper scripts in `scripts/` implement this pattern for public demonstrations.

## Raw prompt handling

The Governance API intentionally does not persist raw prompt content in its audit tables. It stores governance metadata, a prompt hash and prompt length instead.

This reduces retained prompt data, but it is not a complete data-loss-prevention solution. Production deployments should evaluate logging, provider retention, observability pipelines, content filtering and organizational data-handling requirements end to end.

## Production hardening

The current environment is optimized for a low-cost demo. Production deployments should additionally evaluate:

- private networking and controlled egress;
- WAF/API protection and rate limiting;
- SIEM/security monitoring;
- credential rotation automation;
- managed identity scope review;
- database HA/DR and backup policy;
- provider/model safety controls;
- data residency and regulatory requirements;
- CI/CD security gates and artifact signing;
- strict financial reservation/ledger semantics when atomic spend caps are required.
