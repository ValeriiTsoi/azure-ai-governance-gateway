data "azurerm_client_config" "current" {}

locals {
  common_tags = {
    project     = "azure-ai-governance-gateway"
    environment = var.environment
    purpose     = "governance-gateway-demo"
    managed_by  = "terraform"
  }

  tfstate_tags = {
    project    = "azure-ai-governance-gateway"
    purpose    = "terraform-state"
    managed_by = "terraform"
  }

  # Azure Storage Account names must be globally unique.
  # The suffix is deterministic for this Azure subscription.
  tfstate_storage_account_name = "staigovtf${substr(md5(data.azurerm_client_config.current.subscription_id), 0, 8)}"
}

# Main demo resource group
resource "azurerm_resource_group" "demo" {
  name     = "rg-aigov-demo"
  location = var.location

  tags = local.common_tags
}

# Dedicated resource group for Terraform remote state
resource "azurerm_resource_group" "tfstate" {
  name     = "rg-aigov-tfstate"
  location = var.location

  tags = local.tfstate_tags
}

# Storage account for Terraform remote state
resource "azurerm_storage_account" "tfstate" {
  name                     = local.tfstate_storage_account_name
  resource_group_name      = azurerm_resource_group.tfstate.name
  location                 = azurerm_resource_group.tfstate.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"

  allow_nested_items_to_be_public = false

  blob_properties {
    versioning_enabled = true
  }

  tags = local.tfstate_tags
}

# Private Blob container for Terraform state files
resource "azurerm_storage_container" "tfstate" {
  name                  = "tfstate"
  storage_account_id    = azurerm_storage_account.tfstate.id
  container_access_type = "private"
}

# Allow the currently authenticated Azure user to access
# the state blob using Microsoft Entra ID rather than storage keys.
resource "azurerm_role_assignment" "tfstate_blob_contributor" {
  scope                = azurerm_storage_account.tfstate.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = data.azurerm_client_config.current.object_id
}
