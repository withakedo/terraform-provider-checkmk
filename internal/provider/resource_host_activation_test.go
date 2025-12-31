package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostResource_WithAutoActivate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with auto-activation enabled
			{
				Config: testAccHostResourceConfigWithAutoActivate("test-autoactivate-host", "Auto Activate Test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_host.test", "host_name", "test-autoactivate-host"),
					resource.TestCheckResourceAttr("checkmk_host.test", "folder", "/"),
					// With activation, attributes should be immediately available
					resource.TestCheckResourceAttr("checkmk_host.test", "attributes.alias", "Auto Activate Test"),
					resource.TestCheckResourceAttr("checkmk_host.test", "attributes.ipaddress", "127.0.0.1"),
				),
			},
			// Update with auto-activation
			{
				Config: testAccHostResourceConfigWithAutoActivate("test-autoactivate-host", "Updated Auto Activate Test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_host.test", "attributes.alias", "Updated Auto Activate Test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccHostResourceConfigWithAutoActivate(hostName, alias string) string {
	return fmt.Sprintf(`
provider "checkmk" {
  activate = "auto"
}

resource "checkmk_host" "test" {
  host_name = %[1]q
  folder    = "/"

  attributes = {
    alias     = %[2]q
    ipaddress = "127.0.0.1"
  }
}
`, hostName, alias)
}
