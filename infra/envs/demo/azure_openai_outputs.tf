output "azure_openai_name" {
  value = azurerm_cognitive_account.openai.name
}

output "azure_openai_endpoint" {
  value = azurerm_cognitive_account.openai.endpoint
}

output "azure_openai_deployment_name" {
  value = azurerm_cognitive_deployment.gpt_5_mini.name
}

output "azure_openai_model" {
  value = local.azure_openai_model_name
}

output "azure_openai_model_version" {
  value = local.azure_openai_model_version
}
