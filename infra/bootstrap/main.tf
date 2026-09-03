locals {
  common_tags = {
    project     = "azure-ai-governance-gateway"
    environment = var.environment
    purpose     = "sandvik-demo"
    managed_by  = "terraform"
  }
}

resource "azurerm_resource_group" "demo" {
  name     = "rg-aigov-demo"
  location = var.location

  tags = local.common_tags
}
