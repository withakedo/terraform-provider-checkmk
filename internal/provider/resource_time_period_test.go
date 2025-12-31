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

func TestAccTimePeriodResource(t *testing.T) {
	// Time periods require activation

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTimePeriodDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTimePeriodResourceConfig("tf_acc_test_tp", "Test Time Period"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_time_period.test", "name", "tf_acc_test_tp"),
					resource.TestCheckResourceAttr("checkmk_time_period.test", "alias", "Test Time Period"),
					resource.TestCheckResourceAttr("checkmk_time_period.test", "active_time_ranges.#", "5"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "checkmk_time_period.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccTimePeriodResourceConfigUpdated("tf_acc_test_tp", "Updated Time Period"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_time_period.test", "name", "tf_acc_test_tp"),
					resource.TestCheckResourceAttr("checkmk_time_period.test", "alias", "Updated Time Period"),
					resource.TestCheckResourceAttr("checkmk_time_period.test", "active_time_ranges.#", "7"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testAccCheckTimePeriodDestroy verifies that time periods have been destroyed
func testAccCheckTimePeriodDestroy(s *terraform.State) error {
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
		if rs.Type != "checkmk_time_period" {
			continue
		}

		timePeriodName := rs.Primary.ID

		_, err := c.GetTimePeriod(ctx, timePeriodName)
		if err == nil {
			return fmt.Errorf("time period still exists: %s", timePeriodName)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted time period %s, got: %v", timePeriodName, err)
			}
		}
	}

	return nil
}

// testAccTimePeriodResourceConfig generates a basic time period with weekday coverage
func testAccTimePeriodResourceConfig(name, alias string) string {
	return fmt.Sprintf(`
resource "checkmk_time_period" "test" {
  name  = %[1]q
  alias = %[2]q

  active_time_ranges = [
    {
      day = "monday"
      time_ranges = [
        { start = "09:00", end = "17:00" }
      ]
    },
    {
      day = "tuesday"
      time_ranges = [
        { start = "09:00", end = "17:00" }
      ]
    },
    {
      day = "wednesday"
      time_ranges = [
        { start = "09:00", end = "17:00" }
      ]
    },
    {
      day = "thursday"
      time_ranges = [
        { start = "09:00", end = "17:00" }
      ]
    },
    {
      day = "friday"
      time_ranges = [
        { start = "09:00", end = "17:00" }
      ]
    },
  ]
}
`, name, alias)
}

// testAccTimePeriodResourceConfigUpdated generates an updated time period with full week coverage
func testAccTimePeriodResourceConfigUpdated(name, alias string) string {
	return fmt.Sprintf(`
resource "checkmk_time_period" "test" {
  name  = %[1]q
  alias = %[2]q

  active_time_ranges = [
    {
      day = "monday"
      time_ranges = [
        { start = "08:00", end = "18:00" }
      ]
    },
    {
      day = "tuesday"
      time_ranges = [
        { start = "08:00", end = "18:00" }
      ]
    },
    {
      day = "wednesday"
      time_ranges = [
        { start = "08:00", end = "18:00" }
      ]
    },
    {
      day = "thursday"
      time_ranges = [
        { start = "08:00", end = "18:00" }
      ]
    },
    {
      day = "friday"
      time_ranges = [
        { start = "08:00", end = "18:00" }
      ]
    },
    {
      day = "saturday"
      time_ranges = [
        { start = "10:00", end = "14:00" }
      ]
    },
    {
      day = "sunday"
      time_ranges = [
        { start = "10:00", end = "14:00" }
      ]
    },
  ]
}
`, name, alias)
}
