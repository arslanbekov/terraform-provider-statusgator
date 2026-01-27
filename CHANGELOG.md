# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-01-27

### Added

- Initial release of the StatusGator Terraform provider
- Provider configuration with API token authentication
- Resources:
  - `statusgator_monitor_group` - Monitor group management
  - `statusgator_website_monitor` - Website HTTP monitoring
  - `statusgator_ping_monitor` - ICMP ping monitoring
  - `statusgator_custom_monitor` - Custom monitors with manual status
  - `statusgator_service_monitor` - External service status page monitoring
  - `statusgator_subscriber` - Status page subscriber management
  - `statusgator_incident` - Incident and maintenance window management
- Data sources:
  - `statusgator_board` - Get single board details
  - `statusgator_boards` - List all boards
- Import support for all resources
- Full documentation and examples
