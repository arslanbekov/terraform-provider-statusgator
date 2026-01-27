# Terraform Provider for StatusGator

[![Tests](https://github.com/arslanbekov/terraform-provider-statusgator/workflows/Tests/badge.svg)](https://github.com/arslanbekov/terraform-provider-statusgator/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/arslanbekov/terraform-provider-statusgator)](https://goreportcard.com/report/github.com/arslanbekov/terraform-provider-statusgator)

Terraform provider for managing [StatusGator](https://statusgator.com) resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25

## Installation

```terraform
terraform {
  required_providers {
    statusgator = {
      source  = "arslanbekov/statusgator"
      version = "~> 1.0"
    }
  }
}
```

## Configuration

```terraform
provider "statusgator" {
  api_token = var.statusgator_api_token
}
```

Or use environment variables:

```shell
export STATUSGATOR_API_TOKEN="your-api-token"
```

## Resources

- `statusgator_monitor_group` - Manage monitor groups
- `statusgator_website_monitor` - HTTP/HTTPS website monitors
- `statusgator_ping_monitor` - ICMP ping monitors
- `statusgator_custom_monitor` - Manually-controlled monitors
- `statusgator_service_monitor` - External status page monitors
- `statusgator_subscriber` - Email subscribers
- `statusgator_incident` - Incidents and maintenance windows

## Data Sources

- `statusgator_board` - Get board details
- `statusgator_boards` - List all boards

## Quick Start

```terraform
# Create a monitor group
resource "statusgator_monitor_group" "infrastructure" {
  board_id = "my-board-id"
  name     = "Infrastructure"
}

# Monitor a website
resource "statusgator_website_monitor" "api" {
  board_id       = "my-board-id"
  name           = "API Health"
  url            = "https://api.example.com/health"
  check_interval = 5
  group_id       = statusgator_monitor_group.infrastructure.id
}

# Track GitHub status
resource "statusgator_service_monitor" "github" {
  board_id   = "my-board-id"
  service_id = "github"
  name       = "GitHub"
}

# Add a subscriber
resource "statusgator_subscriber" "admin" {
  board_id          = "my-board-id"
  email             = "admin@example.com"
  skip_confirmation = true
}
```

## Development

### Build

```shell
make build
```

### Test

```shell
make test
```

### Acceptance Tests

```shell
export STATUSGATOR_API_TOKEN="your-api-token"
export STATUSGATOR_TEST_BOARD_ID="your-test-board-id"
make testacc
```

### Install Locally

```shell
make install
```

## Documentation

Full documentation is available at [Terraform Registry](https://registry.terraform.io/providers/arslanbekov/statusgator/latest/docs).

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) file for details
