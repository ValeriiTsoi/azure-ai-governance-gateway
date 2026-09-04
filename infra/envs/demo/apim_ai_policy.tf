resource "azurerm_api_management_api_policy" "ai_auth" {
  api_name            = azurerm_api_management_api.ai.name
  api_management_name = azurerm_api_management.demo.name
  resource_group_name = data.azurerm_resource_group.demo.name

  xml_content = templatefile(
    "${path.module}/policies/governance-auth.xml.tftpl",
    {
      tenant_id                = var.entra_tenant_id
      governance_api_client_id = var.governance_api_client_id
      demo_client_id           = var.demo_client_id
    }
  )
}
