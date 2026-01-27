# Create a monitor group
resource "statusgator_monitor_group" "infrastructure" {
  board_id  = "my-board-id"
  name      = "Infrastructure"
  position  = 1
  collapsed = false
}

# Create another group that's collapsed by default
resource "statusgator_monitor_group" "third_party" {
  board_id  = "my-board-id"
  name      = "Third-Party Services"
  position  = 2
  collapsed = true
}
