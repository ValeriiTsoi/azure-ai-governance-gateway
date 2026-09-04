resource "azurerm_api_management_api" "ai" {
  name                = "ai-invocation-api"
  resource_group_name = data.azurerm_resource_group.demo.name
  api_management_name = azurerm_api_management.demo.name

  revision     = "1"
  display_name = "Governed AI Invocation API"
  description  = "Governed AI model invocation API for the Azure AI Governance Gateway demo."

  path      = "v1/ai"
  protocols = ["https"]

  service_url = "https://${azurerm_container_app.governance_api.ingress[0].fqdn}/v1/ai"

  subscription_required = false
}

resource "azurerm_api_management_api_operation" "invoke_ai" {
  operation_id        = "invoke-governed-ai"
  api_name            = azurerm_api_management_api.ai.name
  api_management_name = azurerm_api_management.demo.name
  resource_group_name = data.azurerm_resource_group.demo.name

  display_name = "Invoke Governed AI"
  method       = "POST"
  url_template = "/invoke"

  description = "Evaluates governance policy, routes an allowed request, invokes the configured AI provider, and records usage."
}
