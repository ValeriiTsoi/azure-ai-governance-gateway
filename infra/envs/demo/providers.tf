provider "azurerm" {
  features {}

  resource_provider_registrations = "none"
}

provider "azapi" {
  skip_provider_registration = true
}
