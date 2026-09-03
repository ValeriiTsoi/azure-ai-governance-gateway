output "resource_group_name" {
  description = "Demo resource group name."
  value       = azurerm_resource_group.demo.name
}

output "resource_group_location" {
  description = "Demo resource group Azure region."
  value       = azurerm_resource_group.demo.location
}

output "tfstate_resource_group_name" {
  description = "Terraform state resource group."
  value       = azurerm_resource_group.tfstate.name
}

output "tfstate_storage_account_name" {
  description = "Terraform remote state storage account."
  value       = azurerm_storage_account.tfstate.name
}

output "tfstate_container_name" {
  description = "Terraform state Blob container."
  value       = azurerm_storage_container.tfstate.name
}
