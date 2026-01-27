# Create an incident
resource "statusgator_incident" "outage" {
  board_id = "my-board-id"
  name     = "API Outage"
  details  = "We are investigating reports of API connectivity issues."
  severity = "major"
  phase    = "investigating"
}

# Scheduled maintenance window
resource "statusgator_incident" "maintenance" {
  board_id      = "my-board-id"
  name          = "Scheduled Database Maintenance"
  details       = "We will be performing scheduled maintenance on the database."
  severity      = "maintenance"
  phase         = "scheduled"
  will_start_at = "2025-01-30T02:00:00Z"
  will_end_at   = "2025-01-30T04:00:00Z"
}
