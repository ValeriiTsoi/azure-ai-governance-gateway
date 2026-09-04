locals {
  governance_api_identifier_uri = "api://aigov-governance-api-demo"

  # Stable UUID for the delegated OAuth2 scope.
  access_as_user_scope_id = "a86d7a21-2cbc-4f23-9514-2c6d6e0ced31"
}

resource "azuread_application" "governance_api" {
  display_name = "aigov-governance-api-demo"

  description = "Azure AI Governance Gateway demo API."

  sign_in_audience = "AzureADMyOrg"

  identifier_uris = [
    local.governance_api_identifier_uri
  ]

  owners = [
    data.azuread_client_config.current.object_id
  ]

  api {
    requested_access_token_version = 2

    oauth2_permission_scope {
      admin_consent_description  = "Allow the application to access the AI Governance Gateway on behalf of the signed-in user."
      admin_consent_display_name = "Access AI Governance Gateway"

      enabled = true
      id      = local.access_as_user_scope_id
      type    = "User"

      user_consent_description  = "Allow this application to access the AI Governance Gateway on your behalf."
      user_consent_display_name = "Access AI Governance Gateway"

      value = "access_as_user"
    }
  }
}

resource "azuread_service_principal" "governance_api" {
  client_id = azuread_application.governance_api.client_id

  app_role_assignment_required = false

  owners = [
    data.azuread_client_config.current.object_id
  ]
}
