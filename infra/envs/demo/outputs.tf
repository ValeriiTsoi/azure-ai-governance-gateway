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
