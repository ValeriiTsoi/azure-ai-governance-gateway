resource "azurerm_container_app_environment" "demo" {
  name                = "cae-aigov-demo"
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name

  logs_destination           = "log-analytics"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.demo.id

  # Explicitly keep the Consumption workload profile used by the demo.
  # This prevents Terraform drift after Azure materializes the default profile.
  workload_profile {
    name                  = "Consumption"
    workload_profile_type = "Consumption"
    minimum_count         = 0
    maximum_count         = 0
  }

  tags = local.common_tags
}
