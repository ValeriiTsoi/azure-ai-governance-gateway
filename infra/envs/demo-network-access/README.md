# Demo Network Access

This Terraform root manages the demo-specific PostgreSQL firewall allowlist used by the Azure Container Apps Governance API.

It is deliberately isolated from the main `demo` Terraform state because the Container Apps Consumption environment currently exposes a large outbound IP set. Managing these firewall rules in the main state caused every application infrastructure plan to refresh more than 160 individual PostgreSQL firewall resources.

## State isolation

Main infrastructure:

~~~text
demo.tfstate
~~~

Network access:

~~~text
demo-network-access.tfstate
~~~

The separation allows normal application and platform changes to avoid refreshing the complete PostgreSQL firewall allowlist.

## Backend initialization

~~~bash
terraform init \
  -reconfigure \
  -backend-config="resource_group_name=rg-aigov-tfstate" \
  -backend-config="storage_account_name=staigovtf0d60fe3d" \
  -backend-config="container_name=tfstate" \
  -backend-config="key=demo-network-access.tfstate"
~~~

## Usage

Network-access changes should be managed only from this Terraform root:

~~~bash
terraform fmt
terraform validate
terraform plan
terraform apply
~~~

Do not manage these firewall rules from `infra/envs/demo`.

## Demo networking note

The current Azure Container Apps Consumption environment exposes a relatively large outbound IP set.

The PostgreSQL firewall therefore contains an explicit allowlist for those outbound addresses. Keeping the rules in their own Terraform state prevents ordinary application and infrastructure changes from refreshing more than 160 firewall resources.

## Production architecture

This outbound-IP allowlist is a cost-conscious demo workaround and is not the intended production networking design.

A production deployment should use controlled/private networking and stable egress, for example:

~~~text
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
~~~

The production topology should avoid maintaining a large dynamic Container Apps Consumption outbound-IP allowlist.
