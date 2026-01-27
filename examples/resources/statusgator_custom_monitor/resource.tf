# Custom monitor with manual status control
resource "statusgator_custom_monitor" "deployment" {
  board_id    = "my-board-id"
  name        = "Deployment Pipeline"
  description = "Status of the CI/CD deployment pipeline"
  status      = "up"
}

# Custom monitor in a group
resource "statusgator_custom_monitor" "batch_jobs" {
  board_id    = "my-board-id"
  name        = "Nightly Batch Jobs"
  description = "Status of nightly data processing jobs"
  status      = "up"
  group_id    = statusgator_monitor_group.infrastructure.id
}
