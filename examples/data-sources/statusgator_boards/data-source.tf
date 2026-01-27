# List all boards in the account
data "statusgator_boards" "all" {}

output "board_count" {
  value = length(data.statusgator_boards.all.boards)
}

output "board_names" {
  value = [for board in data.statusgator_boards.all.boards : board.name]
}
