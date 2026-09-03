output "resource_group_name" {
  description = "Demo resource group name."
  value       = azurerm_resource_group.demo.name
}

output "resource_group_location" {
  description = "Demo resource group Azure region."
  value       = azurerm_resource_group.demo.location
}
