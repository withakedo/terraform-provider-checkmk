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

func TestAccAuxTagResource(t *testing.T) {
	// Aux tags do NOT require activation

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAuxTagDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAuxTagResourceConfig("tf_acc_test_aux", "Test Aux Tag"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_aux_tag.test", "id", "tf_acc_test_aux"),
					resource.TestCheckResourceAttr("checkmk_aux_tag.test", "title", "Test Aux Tag"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_aux_tag.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking"},
			},
			// Update and Read testing
			{
				Config: testAccAuxTagResourceConfigWithTopic("tf_acc_test_aux", "Updated Aux Tag", "Custom Tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_aux_tag.test", "id", "tf_acc_test_aux"),
					resource.TestCheckResourceAttr("checkmk_aux_tag.test", "title", "Updated Aux Tag"),
					resource.TestCheckResourceAttr("checkmk_aux_tag.test", "topic", "Custom Tags"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testAccCheckAuxTagDestroy verifies that aux tags have been destroyed
func testAccCheckAuxTagDestroy(s *terraform.State) error {
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
		if rs.Type != "checkmk_aux_tag" {
			continue
		}

		auxTagID := rs.Primary.ID

		_, err := c.GetAuxTag(ctx, auxTagID)
		if err == nil {
			return fmt.Errorf("aux tag still exists: %s", auxTagID)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted aux tag %s, got: %v", auxTagID, err)
			}
		}
	}

	return nil
}

func testAccAuxTagResourceConfig(id, title string) string {
	return fmt.Sprintf(`
resource "checkmk_aux_tag" "test" {
  id    = %[1]q
  title = %[2]q
}
`, id, title)
}

func testAccAuxTagResourceConfigWithTopic(id, title, topic string) string {
	return fmt.Sprintf(`
resource "checkmk_aux_tag" "test" {
  id    = %[1]q
  title = %[2]q
  topic = %[3]q
}
`, id, title, topic)
}
