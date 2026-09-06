resource "azurerm_api_management_api" "openai_compat" {
  name                = "openai-compatible-api"
  resource_group_name = data.azurerm_resource_group.demo.name
  api_management_name = azurerm_api_management.demo.name

  revision     = "1"
  display_name = "OpenAI-Compatible Governed AI API"
  description  = "OpenAI-compatible facade for governed AI requests."

  path      = "openai/v1"
  protocols = ["https"]

  service_url = "https://${azurerm_container_app.governance_api.ingress[0].fqdn}/v1"

  subscription_required = false
}

resource "azurerm_api_management_api_operation" "openai_models" {
  operation_id        = "list-openai-models"
  api_name            = azurerm_api_management_api.openai_compat.name
  api_management_name = azurerm_api_management.demo.name
  resource_group_name = data.azurerm_resource_group.demo.name

  display_name = "List OpenAI-Compatible Models"
  method       = "GET"
  url_template = "/models"

  description = "Lists logical AI models exposed by the governance gateway."
}

resource "azurerm_api_management_api_operation" "openai_chat_completions" {
  operation_id        = "create-openai-chat-completion"
  api_name            = azurerm_api_management_api.openai_compat.name
  api_management_name = azurerm_api_management.demo.name
  resource_group_name = data.azurerm_resource_group.demo.name

  display_name = "Create OpenAI-Compatible Chat Completion"
  method       = "POST"
  url_template = "/chat/completions"

  description = "Invokes the governed AI pipeline using an OpenAI-compatible chat completions request."
}