terraform {
  required_version = ">= 1.16.0, < 2.0.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "= 5.2.0"
    }

    azapi = {
      source  = "Azure/azapi"
      version = "= 2.12.0"
    }
  }
}
