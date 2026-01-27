# Ping monitor for server health
resource "statusgator_ping_monitor" "server1" {
  board_id       = "my-board-id"
  name           = "Server 1"
  host           = "server1.example.com"
  check_interval = 1
  regions        = ["us-east", "us-west"]
}

# Ping monitor in a group
resource "statusgator_ping_monitor" "database" {
  board_id       = "my-board-id"
  name           = "Database Server"
  host           = "db.internal.example.com"
  check_interval = 5
  group_id       = statusgator_monitor_group.infrastructure.id
}
