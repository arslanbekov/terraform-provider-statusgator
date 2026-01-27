package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBoardDataSource(t *testing.T) {
	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBoardDataSourceConfig(boardID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.statusgator_board.test", "id", boardID),
					resource.TestCheckResourceAttrSet("data.statusgator_board.test", "name"),
					resource.TestCheckResourceAttrSet("data.statusgator_board.test", "public_token"),
				),
			},
		},
	})
}

func testAccBoardDataSourceConfig(boardID string) string {
	return fmt.Sprintf(`
data "statusgator_board" "test" {
  id = %[1]q
}
`, boardID)
}
