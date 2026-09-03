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
