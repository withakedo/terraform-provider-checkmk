package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig("testuser", "Test User", "testuser@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_user.test", "username", "testuser"),
					resource.TestCheckResourceAttr("checkmk_user.test", "fullname", "Test User"),
					resource.TestCheckResourceAttr("checkmk_user.test", "email", "testuser@example.com"),
					resource.TestCheckResourceAttr("checkmk_user.test", "id", "testuser"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking", "password", "automation_secret", "enforce_password_change"},
			},
			// Update testing
			{
				Config: testAccUserResourceConfig("testuser", "Updated User", "updated@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_user.test", "username", "testuser"),
					resource.TestCheckResourceAttr("checkmk_user.test", "fullname", "Updated User"),
					resource.TestCheckResourceAttr("checkmk_user.test", "email", "updated@example.com"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccUserResource_WithRoles(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfigWithRoles("testuser_roles", "Test User with Roles"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_user.test", "username", "testuser_roles"),
					resource.TestCheckResourceAttr("checkmk_user.test", "fullname", "Test User with Roles"),
				),
			},
		},
	})
}

func testAccUserResourceConfig(username, fullname, email string) string {
	return fmt.Sprintf(`
resource "checkmk_user" "test" {
  username = %[1]q
  fullname = %[2]q
  email    = %[3]q

  auth_type = "password"
  password  = "TestPassword123!"
}
`, username, fullname, email)
}

func testAccUserResourceConfigWithRoles(username, fullname string) string {
	return fmt.Sprintf(`
resource "checkmk_user" "test" {
  username = %[1]q
  fullname = %[2]q

  roles = ["user"]

  auth_type = "password"
  password  = "TestPassword123!"
}
`, username, fullname)
}
