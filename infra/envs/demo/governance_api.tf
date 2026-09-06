locals {
  governance_api_image = "${azurerm_container_registry.demo.login_server}/governance-api@sha256:00e94f27599810ca99fdfb90d45055e219b3563214d6e8e6a7bd59b299e7f4f7"
}

resource "azurerm_user_assigned_identity" "governance_api" {
  name                = "id-governance-api-demo"
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name

  tags = local.common_tags
}

resource "azurerm_role_assignment" "governance_api_acr_pull" {
  scope                = azurerm_container_registry.demo.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.governance_api.principal_id
  principal_type       = "ServicePrincipal"

  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "governance_api_keyvault_secrets_user" {
  scope                = azurerm_key_vault.demo.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.governance_api.principal_id
  principal_type       = "ServicePrincipal"

  skip_service_principal_aad_check = true
}

resource "azurerm_key_vault_secret" "governance_api_database_url" {
  name = "governance-api-database-url"

  value_wo = "postgres://aigovadmin:${var.postgresql_admin_password}@${azurerm_postgresql_flexible_server.demo.fqdn}:5432/${azurerm_postgresql_flexible_server_database.aigov.name}?sslmode=require"

  value_wo_version = 1
  key_vault_id     = azurerm_key_vault.demo.id

  content_type = "text/plain"
  tags         = local.common_tags

  depends_on = [
    azurerm_role_assignment.current_user_keyvault_secrets
  ]
}

resource "azurerm_container_app" "governance_api" {
  name                         = "ca-governance-api-demo"
  container_app_environment_id = azurerm_container_app_environment.demo.id
  resource_group_name          = data.azurerm_resource_group.demo.name

  revision_mode         = "Single"
  workload_profile_name = "Consumption"

  identity {
    type = "UserAssigned"

    identity_ids = [
      azurerm_user_assigned_identity.governance_api.id
    ]
  }

  registry {
    server   = azurerm_container_registry.demo.login_server
    identity = azurerm_user_assigned_identity.governance_api.id
  }

  secret {
    name                = "database-url"
    key_vault_secret_id = azurerm_key_vault_secret.governance_api_database_url.versionless_id
    identity            = azurerm_user_assigned_identity.governance_api.id
  }

  ingress {
    external_enabled           = true
    target_port                = 8080
    transport                  = "auto"
    allow_insecure_connections = false

    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    min_replicas = 0
    max_replicas = 1

    container {
      name   = "governance-api"
      image  = local.governance_api_image
      cpu    = 0.25
      memory = "0.5Gi"

      env {
        name  = "PORT"
        value = "8080"
      }

      env {
        name        = "DATABASE_URL"
        secret_name = "database-url"
      }

      env {
        name  = "AI_PROVIDER"
        value = "azure-openai"
      }

      env {
        name  = "AZURE_OPENAI_ENDPOINT"
        value = azurerm_cognitive_account.openai.endpoint
      }

      env {
        name  = "AZURE_OPENAI_DEPLOYMENT"
        value = azurerm_cognitive_deployment.gpt_5_mini.name
      }

      env {
        name  = "AZURE_CLIENT_ID"
        value = azurerm_user_assigned_identity.governance_api.client_id
      }

      liveness_probe {
        transport               = "HTTP"
        port                    = 8080
        path                    = "/healthz"
        initial_delay           = 5
        interval_seconds        = 30
        timeout                 = 2
        failure_count_threshold = 3
      }

      readiness_probe {
        transport               = "HTTP"
        port                    = 8080
        path                    = "/readyz"
        initial_delay           = 5
        interval_seconds        = 10
        timeout                 = 2
        failure_count_threshold = 3
      }
    }
  }

  tags = local.common_tags

  depends_on = [
    azurerm_role_assignment.governance_api_acr_pull,
    azurerm_role_assignment.governance_api_keyvault_secrets_user
  ]
}
