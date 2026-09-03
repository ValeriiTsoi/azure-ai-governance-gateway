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
