data "azurerm_subscription" "current" {}

data "azurerm_client_config" "current" {}

data "azurerm_resource_group" "demo" {
  name = "rg-aigov-demo"
}
