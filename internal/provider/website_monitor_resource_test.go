package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccWebsiteMonitorResource(t *testing.T) {
	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccWebsiteMonitorConfig(boardID, "test-website-monitor", "https://example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("statusgator_website_monitor.test", "name", "test-website-monitor"),
					resource.TestCheckResourceAttr("statusgator_website_monitor.test", "url", "https://example.com"),
					resource.TestCheckResourceAttr("statusgator_website_monitor.test", "board_id", boardID),
					resource.TestCheckResourceAttrSet("statusgator_website_monitor.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "statusgator_website_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebsiteMonitorImportStateIdFunc(boardID),
				ImportStateVerifyIgnore: []string{
					"url",
					"check_interval",
					"http_method",
					"expected_status",
					"timeout",
					"follow_redirects",
				},
			},
			// Update
			{
				Config: testAccWebsiteMonitorConfig(boardID, "test-website-monitor-updated", "https://example.com/health"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("statusgator_website_monitor.test", "name", "test-website-monitor-updated"),
					resource.TestCheckResourceAttr("statusgator_website_monitor.test", "url", "https://example.com/health"),
				),
			},
		},
	})
}

func testAccWebsiteMonitorConfig(boardID, name, url string) string {
	return fmt.Sprintf(`
resource "statusgator_website_monitor" "test" {
  board_id        = %[1]q
  name            = %[2]q
  url             = %[3]q
  check_interval  = 5
  http_method     = "GET"
  expected_status = 200
  timeout         = 30
}
`, boardID, name, url)
}

func testAccWebsiteMonitorImportStateIdFunc(boardID string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources["statusgator_website_monitor.test"]
		if !ok {
			return "", fmt.Errorf("resource not found")
		}
		return fmt.Sprintf("%s/%s", boardID, rs.Primary.ID), nil
	}
}
