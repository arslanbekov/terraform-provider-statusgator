# Basic website monitor
resource "statusgator_website_monitor" "api" {
  board_id       = "my-board-id"
  name           = "API Health"
  url            = "https://api.example.com/health"
  check_interval = 5
}

# Website monitor with all options
resource "statusgator_website_monitor" "main_site" {
  board_id         = "my-board-id"
  name             = "Main Website"
  url              = "https://www.example.com"
  check_interval   = 1
  http_method      = "GET"
  expected_status  = 200
  content_match    = "Welcome"
  timeout          = 30
  follow_redirects = true
  group_id         = statusgator_monitor_group.infrastructure.id

  headers = {
    "Authorization" = "Bearer token123"
    "X-Custom"      = "value"
  }

  regions = ["us-east", "eu-west"]
}
