output "subscription_name" {
  value = data.azurerm_subscription.current.display_name
}

output "demo_resource_group_name" {
  value = data.azurerm_resource_group.demo.name
}

output "demo_resource_group_location" {
  value = data.azurerm_resource_group.demo.location
}

output "monthly_budget" {
  value = var.budget_amount
}

output "log_analytics_workspace_name" {
  value = azurerm_log_analytics_workspace.demo.name
}

output "application_insights_name" {
  value = azurerm_application_insights.demo.name
}

output "key_vault_name" {
  value = azurerm_key_vault.demo.name
}

output "container_registry_name" {
  value = azurerm_container_registry.demo.name
}

output "container_registry_login_server" {
  value = azurerm_container_registry.demo.login_server
}

output "container_app_environment_name" {
  value = azurerm_container_app_environment.demo.name
}

output "postgresql_server_name" {
  value = azurerm_postgresql_flexible_server.demo.name
}

output "postgresql_server_fqdn" {
  value = azurerm_postgresql_flexible_server.demo.fqdn
}

output "postgresql_database_name" {
  value = azurerm_postgresql_flexible_server_database.aigov.name
}

output "postgresql_admin_login" {
  value = azurerm_postgresql_flexible_server.demo.administrator_login
}

output "governance_api_identity_name" {
  value = azurerm_user_assigned_identity.governance_api.name
}

output "governance_api_identity_principal_id" {
  value = azurerm_user_assigned_identity.governance_api.principal_id
}

output "governance_api_image" {
  value = local.governance_api_image
}

output "governance_api_fqdn" {
  value = azurerm_container_app.governance_api.ingress[0].fqdn
}

output "governance_api_url" {
  value = "https://${azurerm_container_app.governance_api.ingress[0].fqdn}"
}

output "api_management_name" {
  description = "Name of the demo Azure API Management service."
  value       = azurerm_api_management.demo.name
}

output "api_management_gateway_url" {
  description = "Gateway URL of the demo Azure API Management service."
  value       = azurerm_api_management.demo.gateway_url
}

output "api_management_identity_principal_id" {
  description = "Principal ID of the API Management system-assigned managed identity."
  value       = azurerm_api_management.demo.identity[0].principal_id
}
