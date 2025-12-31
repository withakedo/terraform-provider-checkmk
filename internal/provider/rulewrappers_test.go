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

// TestAccHostCheckIntervalResource tests the checkmk_host_check_interval resource
func TestAccHostCheckIntervalResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_check_interval"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHostCheckIntervalResourceConfig(60, "Test host check interval"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_check_interval.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_host_check_interval.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_host_check_interval.test", "value", "60"),
				),
			},
			// ImportState testing
			{
				ResourceName: "checkmk_host_check_interval.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["checkmk_host_check_interval.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return rs.Primary.Attributes["api_id"], nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"conditions"},
			},
			// Update and Read testing
			{
				Config: testAccHostCheckIntervalResourceConfig(120, "Test host check interval updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_check_interval.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_check_interval.test", "value", "120"),
				),
			},
		},
	})
}

// TestAccServiceCheckIntervalResource tests the checkmk_service_check_interval resource
func TestAccServiceCheckIntervalResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_service_check_interval"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceCheckIntervalResourceConfig(60, "Test service check interval"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_service_check_interval.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_service_check_interval.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_service_check_interval.test", "value", "60"),
				),
			},
			// Update and Read testing
			{
				Config: testAccServiceCheckIntervalResourceConfig(300, "Test service check interval updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_service_check_interval.test", "value", "300"),
				),
			},
		},
	})
}

// TestAccHostNotificationPeriodResource tests the checkmk_host_notification_period resource
func TestAccHostNotificationPeriodResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_notification_period"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHostNotificationPeriodResourceConfig("24X7", "Test host notification period"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_notification_period.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_host_notification_period.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_host_notification_period.test", "value", "24X7"),
				),
			},
			// ImportState testing
			{
				ResourceName: "checkmk_host_notification_period.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["checkmk_host_notification_period.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return rs.Primary.Attributes["api_id"], nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"conditions"},
			},
		},
	})
}

// TestAccServiceNotificationPeriodResource tests the checkmk_service_notification_period resource
func TestAccServiceNotificationPeriodResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_service_notification_period"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceNotificationPeriodResourceConfig("24X7", "Test service notification period"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_service_notification_period.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_service_notification_period.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_service_notification_period.test", "value", "24X7"),
				),
			},
		},
	})
}

// TestAccServiceMaxCheckAttemptsResource tests the checkmk_service_max_check_attempts resource
func TestAccServiceMaxCheckAttemptsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_service_max_check_attempts"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceMaxCheckAttemptsResourceConfig(3, "Test max check attempts"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_service_max_check_attempts.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_service_max_check_attempts.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_service_max_check_attempts.test", "value", "3"),
				),
			},
			// Update and Read testing
			{
				Config: testAccServiceMaxCheckAttemptsResourceConfig(5, "Test max check attempts updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_service_max_check_attempts.test", "value", "5"),
				),
			},
		},
	})
}

// TestAccServiceRetryIntervalResource tests the checkmk_service_retry_interval resource
func TestAccServiceRetryIntervalResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_service_retry_interval"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceRetryIntervalResourceConfig(30, "Test retry interval"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_service_retry_interval.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_service_retry_interval.test", "api_id"),
					resource.TestCheckResourceAttr("checkmk_service_retry_interval.test", "value", "30"),
				),
			},
			// Update and Read testing
			{
				Config: testAccServiceRetryIntervalResourceConfig(60, "Test retry interval updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_service_retry_interval.test", "value", "60"),
				),
			},
		},
	})
}

// testAccCheckRuleWrapperDestroy returns a destroy check function for any rule-based wrapper resource
func testAccCheckRuleWrapperDestroy(resourceType string) func(*terraform.State) error {
	return func(s *terraform.State) error {
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
			if rs.Type != resourceType {
				continue
			}

			apiID := rs.Primary.Attributes["api_id"]

			_, err := c.GetRule(ctx, apiID)
			if err == nil {
				return fmt.Errorf("%s rule still exists: %s", resourceType, apiID)
			}

			if apiErr, ok := err.(*client.APIError); ok {
				if apiErr.Status != 404 {
					return fmt.Errorf("expected 404 error for deleted %s rule %s, got: %v", resourceType, apiID, err)
				}
			}
		}

		return nil
	}
}

// Config generators

func testAccHostCheckIntervalResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_check_interval" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

func testAccServiceCheckIntervalResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_service_check_interval" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

func testAccHostNotificationPeriodResourceConfig(value, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_notification_period" "test" {
  folder = "/"
  value  = %q

  properties = {
    description = %q
  }
}
`, value, description)
}

func testAccServiceNotificationPeriodResourceConfig(value, description string) string {
	return fmt.Sprintf(`
resource "checkmk_service_notification_period" "test" {
  folder = "/"
  value  = %q

  properties = {
    description = %q
  }
}
`, value, description)
}

func testAccServiceMaxCheckAttemptsResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_service_max_check_attempts" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

func testAccServiceRetryIntervalResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_service_retry_interval" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

// Additional wrapper tests

// TestAccHostMaxCheckAttemptsResource tests the checkmk_host_max_check_attempts resource
func TestAccHostMaxCheckAttemptsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_max_check_attempts"),
		Steps: []resource.TestStep{
			{
				Config: testAccHostMaxCheckAttemptsResourceConfig(3, "Test host max check attempts"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_max_check_attempts.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_max_check_attempts.test", "value", "3"),
				),
			},
			{
				Config: testAccHostMaxCheckAttemptsResourceConfig(5, "Test host max check attempts updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_host_max_check_attempts.test", "value", "5"),
				),
			},
		},
	})
}

func testAccHostMaxCheckAttemptsResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_max_check_attempts" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccHostRetryIntervalResource tests the checkmk_host_retry_interval resource
func TestAccHostRetryIntervalResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_retry_interval"),
		Steps: []resource.TestStep{
			{
				Config: testAccHostRetryIntervalResourceConfig(30, "Test host retry interval"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_retry_interval.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_retry_interval.test", "value", "30"),
				),
			},
		},
	})
}

func testAccHostRetryIntervalResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_retry_interval" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccHostCheckPeriodResource tests the checkmk_host_check_period resource
func TestAccHostCheckPeriodResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_check_period"),
		Steps: []resource.TestStep{
			{
				Config: testAccHostCheckPeriodResourceConfig("24X7", "Test host check period"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_check_period.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_check_period.test", "value", "24X7"),
				),
			},
		},
	})
}

func testAccHostCheckPeriodResourceConfig(value, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_check_period" "test" {
  folder = "/"
  value  = %q

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccServiceCheckPeriodResource tests the checkmk_service_check_period resource
func TestAccServiceCheckPeriodResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_service_check_period"),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceCheckPeriodResourceConfig("24X7", "Test service check period"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_service_check_period.test", "id"),
					resource.TestCheckResourceAttr("checkmk_service_check_period.test", "value", "24X7"),
				),
			},
		},
	})
}

func testAccServiceCheckPeriodResourceConfig(value, description string) string {
	return fmt.Sprintf(`
resource "checkmk_service_check_period" "test" {
  folder = "/"
  value  = %q

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccHostNotificationsEnabledResource tests the checkmk_host_notifications_enabled resource
func TestAccHostNotificationsEnabledResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_notifications_enabled"),
		Steps: []resource.TestStep{
			{
				Config: testAccHostNotificationsEnabledResourceConfig(true, "Test host notifications enabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_notifications_enabled.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_notifications_enabled.test", "value", "true"),
				),
			},
			{
				Config: testAccHostNotificationsEnabledResourceConfig(false, "Test host notifications disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_host_notifications_enabled.test", "value", "false"),
				),
			},
		},
	})
}

func testAccHostNotificationsEnabledResourceConfig(value bool, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_notifications_enabled" "test" {
  folder = "/"
  value  = %t

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccServiceActiveChecksEnabledResource tests the checkmk_service_active_checks_enabled resource
func TestAccServiceActiveChecksEnabledResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_service_active_checks_enabled"),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActiveChecksEnabledResourceConfig(true, "Test service active checks enabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_service_active_checks_enabled.test", "id"),
					resource.TestCheckResourceAttr("checkmk_service_active_checks_enabled.test", "value", "true"),
				),
			},
		},
	})
}

func testAccServiceActiveChecksEnabledResourceConfig(value bool, description string) string {
	return fmt.Sprintf(`
resource "checkmk_service_active_checks_enabled" "test" {
  folder = "/"
  value  = %t

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccHostFirstNotificationDelayResource tests the checkmk_host_first_notification_delay resource
func TestAccHostFirstNotificationDelayResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_first_notification_delay"),
		Steps: []resource.TestStep{
			{
				Config: testAccHostFirstNotificationDelayResourceConfig(300, "Test host first notification delay"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_first_notification_delay.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_first_notification_delay.test", "value", "300"),
				),
			},
		},
	})
}

func testAccHostFirstNotificationDelayResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_first_notification_delay" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}

// TestAccHostNotificationIntervalResource tests the checkmk_host_notification_interval resource
func TestAccHostNotificationIntervalResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRuleWrapperDestroy("checkmk_host_notification_interval"),
		Steps: []resource.TestStep{
			{
				Config: testAccHostNotificationIntervalResourceConfig(120, "Test host notification interval"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_host_notification_interval.test", "id"),
					resource.TestCheckResourceAttr("checkmk_host_notification_interval.test", "value", "120"),
				),
			},
		},
	})
}

func testAccHostNotificationIntervalResourceConfig(value int, description string) string {
	return fmt.Sprintf(`
resource "checkmk_host_notification_interval" "test" {
  folder = "/"
  value  = %d

  properties = {
    description = %q
  }
}
`, value, description)
}
