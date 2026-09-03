resource "azurerm_key_vault" "demo" {
  name                = local.key_vault_name
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name
  tenant_id           = data.azurerm_client_config.current.tenant_id

  sku_name = "standard"

  rbac_authorization_enabled = true

  soft_delete_retention_days = 7
  purge_protection_enabled   = false

  public_network_access_enabled = true

  tags = local.common_tags
}

resource "azurerm_role_assignment" "current_user_keyvault_secrets" {
  scope                = azurerm_key_vault.demo.id
  role_definition_name = "Key Vault Secrets Officer"
  principal_id         = data.azurerm_client_config.current.object_id
}
