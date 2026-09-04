resource "azurerm_api_management" "demo" {
  name                = "apim-aigov-0d60fe3d"
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name

  publisher_name  = "Azure AI Governance Gateway"
  publisher_email = var.apim_publisher_email

  sku_name = "Consumption_0"

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "demo"
    managed_by  = "terraform"
    project     = "azure-ai-governance-gateway"
    purpose     = "sandvik-demo"
  }
}
