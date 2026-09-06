# Demo Network Access

This Terraform root manages the demo-specific PostgreSQL firewall allowlist used by the Azure Container Apps Governance API.

It is deliberately isolated from the main `demo` Terraform state because the Container Apps Consumption environment exposes a relatively large outbound-IP set. Managing these rules in the main state made routine application infrastructure plans unnecessarily slow.

## State isolation

Main infrastructure:

```text
demo.tfstate
```

Network access:

```text
demo-network-access.tfstate
```

The separation allows normal application and platform changes to avoid refreshing the complete PostgreSQL firewall allowlist.

## Backend initialization

Retrieve the backend names from the bootstrap Terraform outputs rather than copying state files or credentials:

```bash
TFSTATE_RG="$(terraform -chdir=../../bootstrap output -raw tfstate_resource_group_name)"
TFSTATE_SA="$(terraform -chdir=../../bootstrap output -raw tfstate_storage_account_name)"
TFSTATE_CONTAINER="$(terraform -chdir=../../bootstrap output -raw tfstate_container_name)"
```

Then initialize this root:

```bash
terraform init \
  -reconfigure \
  -backend-config="resource_group_name=$TFSTATE_RG" \
  -backend-config="storage_account_name=$TFSTATE_SA" \
  -backend-config="container_name=$TFSTATE_CONTAINER" \
  -backend-config="key=demo-network-access.tfstate"
```

## Usage

Network-access changes should be managed only from this Terraform root:

```bash
terraform fmt
terraform validate
terraform plan
terraform apply
```

Do not manage these firewall resources from `infra/envs/demo`.

## Demo networking note

The current Azure Container Apps Consumption environment exposes a relatively large outbound-IP set.

The PostgreSQL firewall therefore contains an explicit allowlist for those outbound addresses. Keeping the rules in their own Terraform state prevents ordinary application and infrastructure changes from refreshing the entire allowlist.

For temporary developer/demo access, use the repository helper scripts, which create a rule for the **exact current public IPv4 address** and remove it after use:

```bash
./scripts/demo-db-access-open.sh
./scripts/demo-db-access-close.sh
```

Do not introduce a broad persistent `0.0.0.0` / "Allow Azure services" rule.

## Production architecture

This outbound-IP allowlist is a cost-conscious demo workaround and is not the intended production networking design.

A production deployment should use controlled/private networking and stable egress, for example:

```text
Azure API Management
        |
        v
Container Apps / AKS
        |
        v
VNet
        |
        +---- Private connectivity
        |
        +---- Controlled / stable egress
        |
        v
Azure PostgreSQL
```
