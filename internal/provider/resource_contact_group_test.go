package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-provider-checkmk/internal/client"
)

func TestAccContactGroupResource(t *testing.T) {
	// Contact groups do NOT require activation

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckContactGroupDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContactGroupResourceConfig("tf_acc_test_cg", "Test Contact Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_contact_group.test", "name", "tf_acc_test_cg"),
					resource.TestCheckResourceAttr("checkmk_contact_group.test", "alias", "Test Contact Group"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_contact_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking"},
			},
			// Update and Read testing
			{
				Config: testAccContactGroupResourceConfig("tf_acc_test_cg", "Updated Contact Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_contact_group.test", "name", "tf_acc_test_cg"),
					resource.TestCheckResourceAttr("checkmk_contact_group.test", "alias", "Updated Contact Group"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testAccCheckContactGroupDestroy verifies that contact groups have been destroyed
func testAccCheckContactGroupDestroy(s *terraform.State) error {
	c, err := client.NewClient(
		os.Getenv("CHECKMK_URL"),
		os.Getenv("CHECKMK_USERNAME"),
		os.Getenv("CHECKMK_PASSWORD"),
	)
	if err != nil {
		return fmt.Errorf("failed to create client for destroy check: %s", err)
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "checkmk_contact_group" {
			continue
		}

		contactGroupName := rs.Primary.ID

		_, err := c.GetContactGroup(ctx, contactGroupName)
		if err == nil {
			return fmt.Errorf("contact group still exists: %s", contactGroupName)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted contact group %s, got: %v", contactGroupName, err)
			}
		}
	}

	return nil
}

func testAccContactGroupResourceConfig(name, alias string) string {
	return fmt.Sprintf(`
resource "checkmk_contact_group" "test" {
  name  = %[1]q
  alias = %[2]q
}
`, name, alias)
}
