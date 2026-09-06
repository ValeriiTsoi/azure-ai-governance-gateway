resource "azurerm_role_assignment" "apim_keyvault_secrets_user" {
  scope                = azurerm_key_vault.demo.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_api_management.demo.identity[0].principal_id
  principal_type       = "ServicePrincipal"
}

resource "azurerm_api_management_named_value" "cursor_demo_api_key" {
  name                = "cursor-demo-api-key"
  resource_group_name = data.azurerm_resource_group.demo.name
  api_management_name = azurerm_api_management.demo.name

  display_name = "cursor-demo-api-key"
  secret       = true

  value_from_key_vault {
    secret_id = "${trimsuffix(azurerm_key_vault.demo.vault_uri, "/")}/secrets/cursor-demo-api-key"
  }

  tags = [
    "stage7b",
    "cursor-demo"
  ]

  depends_on = [
    azurerm_role_assignment.apim_keyvault_secrets_user
  ]
}