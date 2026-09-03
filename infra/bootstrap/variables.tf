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

variable "project_name" {
  description = "Short project identifier."
  type        = string
  default     = "aigov"
}
