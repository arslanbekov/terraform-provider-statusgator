terraform {
  required_providers {
    statusgator = {
      source = "arslanbekov/statusgator"
    }
  }
}

# Configure the StatusGator Provider
provider "statusgator" {
  # API token can also be set via STATUSGATOR_API_TOKEN environment variable
  api_token = var.statusgator_api_token

  # Optional: Custom base URL (defaults to https://statusgator.com/api/v3)
  # base_url = "https://statusgator.com/api/v3"

  # Optional: Request timeout in seconds (defaults to 30)
  # timeout = 30
}

variable "statusgator_api_token" {
  description = "StatusGator API token"
  type        = string
  sensitive   = true
}
