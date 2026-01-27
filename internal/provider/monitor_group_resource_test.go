package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccMonitorGroupResource(t *testing.T) {
	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccMonitorGroupConfig(boardID, "test-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("statusgator_monitor_group.test", "name", "test-group"),
					resource.TestCheckResourceAttr("statusgator_monitor_group.test", "board_id", boardID),
					resource.TestCheckResourceAttrSet("statusgator_monitor_group.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "statusgator_monitor_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccMonitorGroupImportStateIdFunc(boardID),
			},
			// Update
			{
				Config: testAccMonitorGroupConfig(boardID, "test-group-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("statusgator_monitor_group.test", "name", "test-group-updated"),
				),
			},
		},
	})
}

func testAccMonitorGroupConfig(boardID, name string) string {
	return fmt.Sprintf(`
resource "statusgator_monitor_group" "test" {
  board_id  = %[1]q
  name      = %[2]q
  position  = 1
  collapsed = false
}
`, boardID, name)
}

func testAccMonitorGroupImportStateIdFunc(boardID string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources["statusgator_monitor_group.test"]
		if !ok {
			return "", fmt.Errorf("resource not found")
		}
		return fmt.Sprintf("%s/%s", boardID, rs.Primary.ID), nil
	}
}
