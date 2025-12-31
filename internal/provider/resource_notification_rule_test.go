package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-provider-checkmk/internal/client"
)

// isCheckmk22 returns true if running against CheckMK 2.2.x
func isCheckmk22() bool {
	url := os.Getenv("CHECKMK_URL")
	return strings.Contains(url, "5020") || strings.Contains(url, "2.2")
}

func TestAccNotificationRuleResource(t *testing.T) {
	// Notification rules require activation

	// Choose config based on CheckMK version
	// CheckMK 2.2 has a typo in the API: "sort_order_for_bulk_notificaions" (missing 't')
	configFunc := testAccNotificationRuleResourceConfig
	updatedConfigFunc := testAccNotificationRuleResourceConfigUpdated
	if isCheckmk22() {
		configFunc = testAccNotificationRuleResourceConfig22
		updatedConfigFunc = testAccNotificationRuleResourceConfigUpdated22
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: configFunc("Test Email Notification"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_notification_rule.test", "id"),
					resource.TestCheckResourceAttrSet("checkmk_notification_rule.test", "rule_config"),
				),
			},
			// ImportState testing
			// Note: rule_config is ignored because JSON key ordering differs between
			// Terraform's jsonencode output and the API's response format
			{
				ResourceName:            "checkmk_notification_rule.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking", "rule_config"},
			},
			// Update and Read testing
			{
				Config: updatedConfigFunc("Updated Email Notification"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("checkmk_notification_rule.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testAccCheckNotificationRuleDestroy verifies that notification rules have been destroyed
func testAccCheckNotificationRuleDestroy(s *terraform.State) error {
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
		if rs.Type != "checkmk_notification_rule" {
			continue
		}

		ruleID := rs.Primary.ID

		_, err := c.GetNotificationRule(ctx, ruleID)
		if err == nil {
			return fmt.Errorf("notification rule still exists: %s", ruleID)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted notification rule %s, got: %v", ruleID, err)
			}
		}
	}

	return nil
}

func testAccNotificationRuleResourceConfig(description string) string {
	return fmt.Sprintf(`
resource "checkmk_notification_rule" "test" {
  rule_config = jsonencode({
    rule_properties = {
      description = %[1]q
      comment = ""
      documentation_url = ""
      do_not_apply_this_rule = { state = "disabled" }
      allow_users_to_deactivate = { state = "enabled" }
    }
    notification_method = {
      notify_plugin = {
        option = "create_notification_with_the_following_parameters"
        plugin_params = {
          plugin_name = "mail"
          from_details = { state = "disabled" }
          reply_to = { state = "disabled" }
          subject_for_host_notifications = { state = "disabled" }
          subject_for_service_notifications = { state = "disabled" }
          send_separate_notification_to_every_recipient = { state = "disabled" }
          sort_order_for_bulk_notifications = { state = "disabled" }
          info_to_be_displayed_in_the_email_body = { state = "disabled" }
          insert_html_section_between_body_and_table = { state = "disabled" }
          url_prefix_for_links_to_checkmk = { state = "disabled" }
          display_graphs_among_each_other = { state = "disabled" }
          enable_sync_smtp = { state = "disabled" }
          graphs_per_notification = { state = "disabled" }
          bulk_notifications_with_graphs = { state = "disabled" }
        }
      }
      notification_bulking = { state = "disabled" }
    }
    contact_selection = {
      all_contacts_of_the_notified_object = { state = "enabled" }
      all_users = { state = "disabled" }
      all_users_with_an_email_address = { state = "disabled" }
      the_following_users = { state = "disabled" }
      members_of_contact_groups = { state = "disabled" }
      explicit_email_addresses = { state = "disabled" }
      restrict_by_custom_macros = { state = "disabled" }
      restrict_by_contact_groups = { state = "disabled" }
    }
    conditions = {
      match_sites = { state = "disabled" }
      match_folder = { state = "disabled" }
      match_host_tags = { state = "disabled" }
      match_host_labels = { state = "disabled" }
      match_host_groups = { state = "disabled" }
      match_hosts = { state = "disabled" }
      match_exclude_hosts = { state = "disabled" }
      match_service_labels = { state = "disabled" }
      match_service_groups = { state = "disabled" }
      match_exclude_service_groups = { state = "disabled" }
      match_service_groups_regex = { state = "disabled" }
      match_exclude_service_groups_regex = { state = "disabled" }
      match_services = { state = "disabled" }
      match_exclude_services = { state = "disabled" }
      match_check_types = { state = "disabled" }
      match_plugin_output = { state = "disabled" }
      match_contact_groups = { state = "disabled" }
      match_service_levels = { state = "disabled" }
      match_only_during_time_period = { state = "disabled" }
      match_host_event_type = { state = "disabled" }
      match_service_event_type = { state = "disabled" }
      restrict_to_notification_numbers = { state = "disabled" }
      throttle_periodic_notifications = { state = "disabled" }
      match_notification_comment = { state = "disabled" }
      event_console_alerts = { state = "disabled" }
    }
  })
}
`, description)
}

func testAccNotificationRuleResourceConfigUpdated(description string) string {
	return fmt.Sprintf(`
resource "checkmk_notification_rule" "test" {
  rule_config = jsonencode({
    rule_properties = {
      description = %[1]q
      comment = "Updated via Terraform"
      documentation_url = ""
      do_not_apply_this_rule = { state = "disabled" }
      allow_users_to_deactivate = { state = "enabled" }
    }
    notification_method = {
      notify_plugin = {
        option = "create_notification_with_the_following_parameters"
        plugin_params = {
          plugin_name = "mail"
          from_details = { state = "disabled" }
          reply_to = { state = "disabled" }
          subject_for_host_notifications = { state = "disabled" }
          subject_for_service_notifications = { state = "disabled" }
          send_separate_notification_to_every_recipient = { state = "disabled" }
          sort_order_for_bulk_notifications = { state = "disabled" }
          info_to_be_displayed_in_the_email_body = { state = "disabled" }
          insert_html_section_between_body_and_table = { state = "disabled" }
          url_prefix_for_links_to_checkmk = { state = "disabled" }
          display_graphs_among_each_other = { state = "disabled" }
          enable_sync_smtp = { state = "disabled" }
          graphs_per_notification = { state = "disabled" }
          bulk_notifications_with_graphs = { state = "disabled" }
        }
      }
      notification_bulking = { state = "disabled" }
    }
    contact_selection = {
      all_contacts_of_the_notified_object = { state = "enabled" }
      all_users = { state = "disabled" }
      all_users_with_an_email_address = { state = "disabled" }
      the_following_users = { state = "disabled" }
      members_of_contact_groups = { state = "disabled" }
      explicit_email_addresses = { state = "disabled" }
      restrict_by_custom_macros = { state = "disabled" }
      restrict_by_contact_groups = { state = "disabled" }
    }
    conditions = {
      match_sites = { state = "disabled" }
      match_folder = { state = "disabled" }
      match_host_tags = { state = "disabled" }
      match_host_labels = { state = "disabled" }
      match_host_groups = { state = "disabled" }
      match_hosts = { state = "disabled" }
      match_exclude_hosts = { state = "disabled" }
      match_service_labels = { state = "disabled" }
      match_service_groups = { state = "disabled" }
      match_exclude_service_groups = { state = "disabled" }
      match_service_groups_regex = { state = "disabled" }
      match_exclude_service_groups_regex = { state = "disabled" }
      match_services = { state = "disabled" }
      match_exclude_services = { state = "disabled" }
      match_check_types = { state = "disabled" }
      match_plugin_output = { state = "disabled" }
      match_contact_groups = { state = "disabled" }
      match_service_levels = { state = "disabled" }
      match_only_during_time_period = { state = "disabled" }
      match_host_event_type = { state = "disabled" }
      match_service_event_type = { state = "disabled" }
      restrict_to_notification_numbers = { state = "disabled" }
      throttle_periodic_notifications = { state = "disabled" }
      match_notification_comment = { state = "disabled" }
      event_console_alerts = { state = "disabled" }
    }
  })
}
`, description)
}

// CheckMK 2.2 has a typo in the API field name: "sort_order_for_bulk_notificaions" (missing 't')
func testAccNotificationRuleResourceConfig22(description string) string {
	return fmt.Sprintf(`
resource "checkmk_notification_rule" "test" {
  rule_config = jsonencode({
    rule_properties = {
      description = %[1]q
      comment = ""
      documentation_url = ""
      do_not_apply_this_rule = { state = "disabled" }
      allow_users_to_deactivate = { state = "enabled" }
    }
    notification_method = {
      notify_plugin = {
        option = "create_notification_with_the_following_parameters"
        plugin_params = {
          plugin_name = "mail"
          from_details = { state = "disabled" }
          reply_to = { state = "disabled" }
          subject_for_host_notifications = { state = "disabled" }
          subject_for_service_notifications = { state = "disabled" }
          send_separate_notification_to_every_recipient = { state = "disabled" }
          sort_order_for_bulk_notificaions = { state = "disabled" }
          info_to_be_displayed_in_the_email_body = { state = "disabled" }
          insert_html_section_between_body_and_table = { state = "disabled" }
          url_prefix_for_links_to_checkmk = { state = "disabled" }
          display_graphs_among_each_other = { state = "disabled" }
          enable_sync_smtp = { state = "disabled" }
          graphs_per_notification = { state = "disabled" }
          bulk_notifications_with_graphs = { state = "disabled" }
        }
      }
      notification_bulking = { state = "disabled" }
    }
    contact_selection = {
      all_contacts_of_the_notified_object = { state = "enabled" }
      all_users = { state = "disabled" }
      all_users_with_an_email_address = { state = "disabled" }
      the_following_users = { state = "disabled" }
      members_of_contact_groups = { state = "disabled" }
      explicit_email_addresses = { state = "disabled" }
      restrict_by_custom_macros = { state = "disabled" }
      restrict_by_contact_groups = { state = "disabled" }
    }
    conditions = {
      match_sites = { state = "disabled" }
      match_folder = { state = "disabled" }
      match_host_tags = { state = "disabled" }
      match_host_labels = { state = "disabled" }
      match_host_groups = { state = "disabled" }
      match_hosts = { state = "disabled" }
      match_exclude_hosts = { state = "disabled" }
      match_service_labels = { state = "disabled" }
      match_service_groups = { state = "disabled" }
      match_exclude_service_groups = { state = "disabled" }
      match_service_groups_regex = { state = "disabled" }
      match_exclude_service_groups_regex = { state = "disabled" }
      match_services = { state = "disabled" }
      match_exclude_services = { state = "disabled" }
      match_check_types = { state = "disabled" }
      match_plugin_output = { state = "disabled" }
      match_contact_groups = { state = "disabled" }
      match_service_levels = { state = "disabled" }
      match_only_during_time_period = { state = "disabled" }
      match_host_event_type = { state = "disabled" }
      match_service_event_type = { state = "disabled" }
      restrict_to_notification_numbers = { state = "disabled" }
      throttle_periodic_notifications = { state = "disabled" }
      match_notification_comment = { state = "disabled" }
      event_console_alerts = { state = "disabled" }
    }
  })
}
`, description)
}

func testAccNotificationRuleResourceConfigUpdated22(description string) string {
	return fmt.Sprintf(`
resource "checkmk_notification_rule" "test" {
  rule_config = jsonencode({
    rule_properties = {
      description = %[1]q
      comment = "Updated via Terraform"
      documentation_url = ""
      do_not_apply_this_rule = { state = "disabled" }
      allow_users_to_deactivate = { state = "enabled" }
    }
    notification_method = {
      notify_plugin = {
        option = "create_notification_with_the_following_parameters"
        plugin_params = {
          plugin_name = "mail"
          from_details = { state = "disabled" }
          reply_to = { state = "disabled" }
          subject_for_host_notifications = { state = "disabled" }
          subject_for_service_notifications = { state = "disabled" }
          send_separate_notification_to_every_recipient = { state = "disabled" }
          sort_order_for_bulk_notificaions = { state = "disabled" }
          info_to_be_displayed_in_the_email_body = { state = "disabled" }
          insert_html_section_between_body_and_table = { state = "disabled" }
          url_prefix_for_links_to_checkmk = { state = "disabled" }
          display_graphs_among_each_other = { state = "disabled" }
          enable_sync_smtp = { state = "disabled" }
          graphs_per_notification = { state = "disabled" }
          bulk_notifications_with_graphs = { state = "disabled" }
        }
      }
      notification_bulking = { state = "disabled" }
    }
    contact_selection = {
      all_contacts_of_the_notified_object = { state = "enabled" }
      all_users = { state = "disabled" }
      all_users_with_an_email_address = { state = "disabled" }
      the_following_users = { state = "disabled" }
      members_of_contact_groups = { state = "disabled" }
      explicit_email_addresses = { state = "disabled" }
      restrict_by_custom_macros = { state = "disabled" }
      restrict_by_contact_groups = { state = "disabled" }
    }
    conditions = {
      match_sites = { state = "disabled" }
      match_folder = { state = "disabled" }
      match_host_tags = { state = "disabled" }
      match_host_labels = { state = "disabled" }
      match_host_groups = { state = "disabled" }
      match_hosts = { state = "disabled" }
      match_exclude_hosts = { state = "disabled" }
      match_service_labels = { state = "disabled" }
      match_service_groups = { state = "disabled" }
      match_exclude_service_groups = { state = "disabled" }
      match_service_groups_regex = { state = "disabled" }
      match_exclude_service_groups_regex = { state = "disabled" }
      match_services = { state = "disabled" }
      match_exclude_services = { state = "disabled" }
      match_check_types = { state = "disabled" }
      match_plugin_output = { state = "disabled" }
      match_contact_groups = { state = "disabled" }
      match_service_levels = { state = "disabled" }
      match_only_during_time_period = { state = "disabled" }
      match_host_event_type = { state = "disabled" }
      match_service_event_type = { state = "disabled" }
      restrict_to_notification_numbers = { state = "disabled" }
      throttle_periodic_notifications = { state = "disabled" }
      match_notification_comment = { state = "disabled" }
      event_console_alerts = { state = "disabled" }
    }
  })
}
`, description)
}
