output "tenant_id" {
  description = "Microsoft Entra tenant ID."
  value       = data.azuread_client_config.current.tenant_id
}

output "governance_api_client_id" {
  description = "Client ID of the Governance API Entra application."
  value       = azuread_application.governance_api.client_id
}

output "governance_api_object_id" {
  description = "Object ID of the Governance API Entra application."
  value       = azuread_application.governance_api.object_id
}

output "governance_api_service_principal_object_id" {
  description = "Object ID of the Governance API service principal."
  value       = azuread_service_principal.governance_api.object_id
}

output "governance_api_identifier_uri" {
  description = "OAuth audience/resource identifier for the Governance API."
  value       = local.governance_api_identifier_uri
}

output "governance_api_scope" {
  description = "Delegated OAuth2 scope used by clients."
  value       = "${local.governance_api_identifier_uri}/access_as_user"
}

output "demo_client_id" {
  description = "Client ID of the public demo client application."
  value       = azuread_application.demo_client.client_id
}

output "demo_client_object_id" {
  description = "Object ID of the public demo client application."
  value       = azuread_application.demo_client.object_id
}

output "demo_client_service_principal_object_id" {
  description = "Object ID of the demo client service principal."
  value       = azuread_service_principal.demo_client.object_id
}
