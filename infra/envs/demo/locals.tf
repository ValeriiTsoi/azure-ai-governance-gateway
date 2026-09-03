locals {
  common_tags = {
    project     = "azure-ai-governance-gateway"
    environment = var.environment
    purpose     = "sandvik-demo"
    managed_by  = "terraform"
  }

  # Key Vault names must be globally unique.
  unique_suffix = substr(
    md5(data.azurerm_client_config.current.subscription_id),
    0,
    8
  )

  key_vault_name = "kv-aigov-${local.unique_suffix}"
}
