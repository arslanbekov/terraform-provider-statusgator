# Get a specific board by ID
data "statusgator_board" "main" {
  id = "my-board-id"
}

output "board_name" {
  value = data.statusgator_board.main.name
}

output "public_url" {
  value = "https://statusgator.com/status/${data.statusgator_board.main.public_token}"
}
