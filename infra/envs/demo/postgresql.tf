variable "postgresql_admin_password" {
  description = "Administrator password for the demo PostgreSQL Flexible Server."
  type        = string
  sensitive   = true
  ephemeral   = true
}

resource "azurerm_postgresql_flexible_server" "demo" {
  name                = "psql-aigov-0d60fe3d"
  resource_group_name = data.azurerm_resource_group.demo.name
  location            = data.azurerm_resource_group.demo.location

  version = "16"

  administrator_login               = "aigovadmin"
  administrator_password_wo         = var.postgresql_admin_password
  administrator_password_wo_version = 1

  sku_name = "B_Standard_B1ms"

  storage_mb        = 32768
  storage_tier      = "P4"
  auto_grow_enabled = false

  backup_retention_days        = 7
  geo_redundant_backup_enabled = false

  # Demo MVP:
  # Keep a public endpoint but no firewall rules yet.
  # Access remains blocked until explicit rules are added.
  public_network_access_enabled = true

  # Azure automatically assigns the Availability Zone for the demo server.
  # Do not let Terraform attempt to move/remove the Azure-managed zone.
  lifecycle {
    ignore_changes = [zone]
  }

  tags = local.common_tags
}

resource "azurerm_postgresql_flexible_server_database" "aigov" {
  name      = "aigov"
  server_id = azurerm_postgresql_flexible_server.demo.id

  charset   = "UTF8"
  collation = "en_US.utf8"
}

resource "azurerm_key_vault_secret" "postgresql_admin_password" {
  name             = "postgresql-admin-password"
  value_wo         = var.postgresql_admin_password
  value_wo_version = 1
  key_vault_id     = azurerm_key_vault.demo.id

  content_type = "text/plain"
  tags         = local.common_tags

  depends_on = [
    azurerm_role_assignment.current_user_keyvault_secrets
  ]
}
