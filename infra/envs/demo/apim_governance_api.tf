resource "azurerm_api_management_api" "governance" {
  name                = "governance-api"
  resource_group_name = data.azurerm_resource_group.demo.name
  api_management_name = azurerm_api_management.demo.name

  revision     = "1"
  display_name = "AI Governance API"
  description  = "Governance decision API for the Azure AI Governance Gateway demo."

  path      = "governance"
  protocols = ["https"]

  service_url = "https://${azurerm_container_app.governance_api.ingress[0].fqdn}/v1/governance"

  # Authentication will be enforced with Microsoft Entra ID
  # in the next stage rather than APIM subscription keys.
  subscription_required = false
}

resource "azurerm_api_management_api_operation" "create_governance_request" {
  operation_id        = "create-governance-request"
  api_name            = azurerm_api_management_api.governance.name
  api_management_name = azurerm_api_management.demo.name
  resource_group_name = data.azurerm_resource_group.demo.name

  display_name = "Create Governance Request"
  method       = "POST"
  url_template = "/requests"

  description = "Creates and evaluates a governance request."
}

resource "azurerm_api_management_api_operation" "get_governance_request" {
  operation_id        = "get-governance-request"
  api_name            = azurerm_api_management_api.governance.name
  api_management_name = azurerm_api_management.demo.name
  resource_group_name = data.azurerm_resource_group.demo.name

  display_name = "Get Governance Request"
  method       = "GET"
  url_template = "/requests/{requestID}"

  description = "Returns a persisted governance request and its latest policy decision."

  template_parameter {
    name     = "requestID"
    type     = "string"
    required = true
  }
}
