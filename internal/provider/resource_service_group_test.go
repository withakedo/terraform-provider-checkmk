package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceGroupResourceConfig("test_service_group", "Test Service Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_service_group.test", "name", "test_service_group"),
					resource.TestCheckResourceAttr("checkmk_service_group.test", "alias", "Test Service Group"),
					resource.TestCheckResourceAttr("checkmk_service_group.test", "id", "test_service_group"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_service_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking"},
			},
			// Update testing
			{
				Config: testAccServiceGroupResourceConfig("test_service_group", "Updated Service Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_service_group.test", "name", "test_service_group"),
					resource.TestCheckResourceAttr("checkmk_service_group.test", "alias", "Updated Service Group"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccServiceGroupResourceConfig(name, alias string) string {
	return fmt.Sprintf(`
resource "checkmk_service_group" "test" {
  name  = %[1]q
  alias = %[2]q
}
`, name, alias)
}
