# Monitor an external service's status page
resource "statusgator_service_monitor" "aws" {
  board_id   = "my-board-id"
  service_id = "amazon-web-services"
  name       = "AWS"
}

# Service monitor in a group
resource "statusgator_service_monitor" "github" {
  board_id   = "my-board-id"
  service_id = "github"
  name       = "GitHub"
  group_id   = statusgator_monitor_group.third_party.id
}
