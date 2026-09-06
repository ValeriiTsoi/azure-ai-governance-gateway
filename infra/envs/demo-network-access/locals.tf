locals {
  resource_group_name = "rg-aigov-demo"

  unique_suffix = substr(
    md5(data.azurerm_subscription.current.subscription_id),
    0,
    8
  )

  postgresql_server_name = "psql-aigov-${local.unique_suffix}"

  postgresql_server_id = "/subscriptions/${data.azurerm_subscription.current.subscription_id}/resourceGroups/${local.resource_group_name}/providers/Microsoft.DBforPostgreSQL/flexibleServers/${local.postgresql_server_name}"
}
