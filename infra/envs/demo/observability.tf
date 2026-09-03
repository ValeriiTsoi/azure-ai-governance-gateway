resource "azurerm_log_analytics_workspace" "demo" {
  name                = "log-aigov-demo"
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name

  sku               = "PerGB2018"
  retention_in_days = 30

  # Demo cost guardrail.
  # Stop billable ingestion if an unexpected telemetry spike occurs.
  daily_quota_gb = 0.5

  tags = local.common_tags
}

resource "azurerm_application_insights" "demo" {
  name                = "appi-aigov-demo"
  location            = data.azurerm_resource_group.demo.location
  resource_group_name = data.azurerm_resource_group.demo.name

  application_type = "web"
  workspace_id     = azurerm_log_analytics_workspace.demo.id

  # Demo cost guardrail.
  daily_data_cap_in_gb                 = 0.5
  daily_data_cap_notifications_enabled = true

  tags = local.common_tags
}
