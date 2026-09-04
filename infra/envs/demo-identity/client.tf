resource "azuread_application" "demo_client" {
  display_name = "aigov-demo-client"

  description = "Public test client for Azure AI Governance Gateway demo authentication."

  sign_in_audience = "AzureADMyOrg"

  fallback_public_client_enabled = true

  owners = [
    data.azuread_client_config.current.object_id
  ]

  required_resource_access {
    resource_app_id = azuread_application.governance_api.client_id

    resource_access {
      id   = local.access_as_user_scope_id
      type = "Scope"
    }
  }
}

resource "azuread_service_principal" "demo_client" {
  client_id = azuread_application.demo_client.client_id

  app_role_assignment_required = false

  owners = [
    data.azuread_client_config.current.object_id
  ]
}
