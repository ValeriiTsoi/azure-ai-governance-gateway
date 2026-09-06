locals {
  common_tags = {
    project     = "azure-ai-governance-gateway"
    environment = var.environment
    purpose     = "governance-gateway-demo"
    managed_by  = "terraform"
  }

  # Globally unique and deterministic suffix for this subscription.
  unique_suffix = substr(
    md5(data.azurerm_client_config.current.subscription_id),
    0,
    8
  )

  key_vault_name         = "kv-aigov-${local.unique_suffix}"
  postgresql_server_name = "psql-aigov-${local.unique_suffix}"
  api_management_name    = "apim-aigov-${local.unique_suffix}"

  # Azure Container Registry names may contain only letters and numbers.
  container_registry_name = "acraigov${local.unique_suffix}"
}
