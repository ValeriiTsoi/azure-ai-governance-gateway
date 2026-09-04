locals {
  resource_group_name    = "rg-aigov-demo"
  postgresql_server_name = "psql-aigov-0d60fe3d"

  postgresql_server_id = "/subscriptions/${data.azurerm_subscription.current.subscription_id}/resourceGroups/${local.resource_group_name}/providers/Microsoft.DBforPostgreSQL/flexibleServers/${local.postgresql_server_name}"
}
