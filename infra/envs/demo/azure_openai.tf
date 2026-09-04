locals {
  azure_openai_name            = "aoai-aigov-0d60fe3d"
  azure_openai_deployment_name = "gpt-5-mini"
  azure_openai_model_name      = "gpt-5-mini"
  azure_openai_model_version   = "2025-08-07"
}

resource "azurerm_cognitive_account" "openai" {
  name                = local.azure_openai_name
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name

  kind     = "OpenAI"
  sku_name = "S0"

  custom_subdomain_name = local.azure_openai_name

  # Demo currently uses public Container Apps egress.
  # Authentication is Entra/RBAC only; account keys are disabled.
  public_network_access_enabled = true
  local_auth_enabled            = false

  tags = local.common_tags
}

resource "azurerm_cognitive_deployment" "gpt_5_mini" {
  name                 = local.azure_openai_deployment_name
  cognitive_account_id = azurerm_cognitive_account.openai.id

  model {
    format  = "OpenAI"
    name    = local.azure_openai_model_name
    version = local.azure_openai_model_version
  }

  sku {
    name     = "GlobalStandard"
    capacity = 10
  }

  # Keep the explicitly tested model version until it reaches retirement.
  version_upgrade_option = "OnceCurrentVersionExpired"
}

resource "azurerm_role_assignment" "governance_api_openai_user" {
  scope                = azurerm_cognitive_account.openai.id
  role_definition_name = "Cognitive Services OpenAI User"
  principal_id         = azurerm_user_assigned_identity.governance_api.principal_id
}
