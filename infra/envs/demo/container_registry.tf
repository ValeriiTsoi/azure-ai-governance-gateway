resource "azurerm_container_registry" "demo" {
  name                = local.container_registry_name
  resource_group_name = data.azurerm_resource_group.demo.name
  location            = data.azurerm_resource_group.demo.location

  sku           = "Basic"
  admin_enabled = false

  public_network_access_enabled = true

  tags = local.common_tags
}
