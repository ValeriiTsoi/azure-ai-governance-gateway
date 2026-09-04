resource "azurerm_postgresql_flexible_server_firewall_rule" "governance_api" {
  for_each = local.governance_api_outbound_ips

  name      = "aca-${replace(each.value, ".", "-")}"
  server_id = azurerm_postgresql_flexible_server.demo.id

  start_ip_address = each.value
  end_ip_address   = each.value
}
