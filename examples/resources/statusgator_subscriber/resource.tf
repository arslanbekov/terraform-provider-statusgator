# Add a subscriber to receive status updates
resource "statusgator_subscriber" "admin" {
  board_id          = "my-board-id"
  email             = "admin@example.com"
  skip_confirmation = true
}

# Subscriber requiring email confirmation
resource "statusgator_subscriber" "user" {
  board_id          = "my-board-id"
  email             = "user@example.com"
  skip_confirmation = false
}
