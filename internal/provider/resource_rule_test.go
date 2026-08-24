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

// NOTE: The rule resource currently has issues with conditions handling that need to be fixed.
// These tests are skipped until that is resolved. The topython() function works correctly
// as verified by the unit tests in function_topython_test.go.

func TestAccRuleResource_WithHostTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			// Create rule with host tag conditions
			{
				Config: testAccRuleResourceConfigWithHostTags("Test rule with host tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_rule.test_tags", "id"),
					resource.TestCheckResourceAttrSet("checkmk_rule.test_tags", "api_id"),
					resource.TestCheckResourceAttr("checkmk_rule.test_tags", "ruleset", "extra_host_conf:check_interval"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCheckRuleDestroy(s *terraform.State) error {
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
		if rs.Type != "checkmk_rule" {
			continue
		}

		apiID := rs.Primary.Attributes["api_id"]

		_, err := c.GetRule(ctx, apiID)
		if err == nil {
			return fmt.Errorf("rule still exists: %s", apiID)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted rule %s, got: %v", apiID, err)
			}
		}
	}

	return nil
}

func testAccRuleResourceConfigWithHostTags(description string) string {
	return fmt.Sprintf(`
resource "checkmk_rule" "test_tags" {
  ruleset = "extra_host_conf:check_interval"
  folder  = "/"

  value_raw = "60.0"

  properties = {
    description = %q
  }

  conditions = {
    host_tags = [
      {
        key      = "criticality"
        operator = "is"
        value    = "prod"
      }
    ]
  }
}
`, description)
}
