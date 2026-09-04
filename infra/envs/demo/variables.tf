variable "location" {
  description = "Primary Azure region for the AI Governance demo."
  type        = string
  default     = "swedencentral"
}

variable "environment" {
  description = "Deployment environment."
  type        = string
  default     = "demo"
}

variable "budget_amount" {
  description = "Monthly Azure budget in the billing currency."
  type        = number
  default     = 150
}

variable "budget_notification_email" {
  description = "Email address used for Azure budget notifications."
  type        = string
  sensitive   = true
}

variable "apim_publisher_email" {
  description = "Publisher email used by the demo Azure API Management service."
  type        = string
}
