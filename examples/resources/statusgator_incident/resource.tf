# Create an incident
resource "statusgator_incident" "outage" {
  board_id = "my-board-id"
  title    = "API Outage"
  message  = "We are investigating reports of API connectivity issues."
  severity = "major"
  phase    = "investigating"

  monitor_ids = [
    statusgator_website_monitor.api.id
  ]
}

# Scheduled maintenance window
resource "statusgator_incident" "maintenance" {
  board_id      = "my-board-id"
  title         = "Scheduled Database Maintenance"
  message       = "We will be performing scheduled maintenance on the database."
  severity      = "maintenance"
  phase         = "scheduled"
  scheduled_for = "2025-01-30T02:00:00Z"
  scheduled_end = "2025-01-30T04:00:00Z"

  monitor_ids = [
    statusgator_ping_monitor.database.id
  ]
}
