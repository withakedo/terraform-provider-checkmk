package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHostGroupResourceConfig("test_host_group", "Test Host Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_host_group.test", "name", "test_host_group"),
					resource.TestCheckResourceAttr("checkmk_host_group.test", "alias", "Test Host Group"),
					resource.TestCheckResourceAttr("checkmk_host_group.test", "id", "test_host_group"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_host_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking"},
			},
			// Update testing
			{
				Config: testAccHostGroupResourceConfig("test_host_group", "Updated Host Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_host_group.test", "name", "test_host_group"),
					resource.TestCheckResourceAttr("checkmk_host_group.test", "alias", "Updated Host Group"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccHostGroupResourceConfig(name, alias string) string {
	return fmt.Sprintf(`
resource "checkmk_host_group" "test" {
  name  = %[1]q
  alias = %[2]q
}
`, name, alias)
}
