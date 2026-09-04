resource "azapi_resource" "governance_api_auth" {
  type      = "Microsoft.App/containerApps/authConfigs@2026-01-01"
  name      = "current"
  parent_id = azurerm_container_app.governance_api.id

  body = {
    properties = {
      platform = {
        enabled = true
      }

      globalValidation = {
        unauthenticatedClientAction = "Return401"

        # Keep operational probes accessible.
        excludedPaths = [
          "/healthz",
          "/readyz"
        ]
      }

      httpSettings = {
        requireHttps = true
      }

      identityProviders = {
        azureActiveDirectory = {
          enabled = true

          registration = {
            clientId     = var.governance_api_client_id
            openIdIssuer = "https://login.microsoftonline.com/${var.entra_tenant_id}/v2.0"
          }

          validation = {
            allowedAudiences = [
              var.governance_api_client_id
            ]

            defaultAuthorizationPolicy = {
              allowedPrincipals = {
                identities = [
                  azurerm_api_management.demo.identity[0].principal_id
                ]
              }
            }
          }
        }
      }

      login = {
        tokenStore = {
          enabled = false
        }
      }
    }
  }
}
