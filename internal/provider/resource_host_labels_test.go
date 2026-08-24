package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/withakedo/terraform_checkmk_provider/internal/client"
)

func TestAccHostLabelsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckHostLabelsDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHostLabelsResourceConfig("Test host labels", map[string]string{"env": "test", "tier": "web"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_labels.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_host_labels.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_host_labels.test", "labels.env", "test"),
					resource.TestCheckResourceAttr("checkmk_host_labels.test", "labels.tier", "web"),
				),
			},
			// ImportState testing
			{
				ResourceName: "checkmk_host_labels.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["checkmk_host_labels.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return rs.Primary.Attributes["api_id"], nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"conditions"}, // Conditions may have default empty values
			},
			// Update and Read testing
			{
				Config: testAccHostLabelsResourceConfig("Test host labels updated", map[string]string{"env": "prod", "tier": "frontend", "region": "us-east"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_labels.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_labels.test", "labels.env", "prod"),
					resource.TestCheckResourceAttr("checkmk_host_labels.test", "labels.tier", "frontend"),
					resource.TestCheckResourceAttr("checkmk_host_labels.test", "labels.region", "us-east"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCheckHostLabelsDestroy(s *terraform.State) error {
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
		if rs.Type != "checkmk_host_labels" {
			continue
		}

		apiID := rs.Primary.Attributes["api_id"]

		_, err := c.GetRule(ctx, apiID)
		if err == nil {
			return fmt.Errorf("host labels rule still exists: %s", apiID)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted host labels rule %s, got: %v", apiID, err)
			}
		}
	}

	return nil
}

func testAccHostLabelsResourceConfig(description string, labels map[string]string) string {
	labelsStr := ""
	for k, v := range labels {
		labelsStr += fmt.Sprintf(`    %s = %q`+"\n", k, v)
	}

	return fmt.Sprintf(`
resource "checkmk_host_labels" "test" {
  folder = "/"

  labels = {
%s  }

  properties = {
    description = %q
  }
}
`, labelsStr, description)
}
