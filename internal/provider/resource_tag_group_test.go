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

func TestAccTagGroupResource(t *testing.T) {
	// Tag groups do NOT require activation - changes take effect immediately

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTagGroupDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTagGroupResourceConfig("tf_acc_test_env", "Test Environment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "id", "tf_acc_test_env"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "title", "Test Environment"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "tags.#", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_tag_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking"},
			},
			// Update and Read testing
			{
				Config: testAccTagGroupResourceConfigUpdated("tf_acc_test_env", "Updated Test Environment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "id", "tf_acc_test_env"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "title", "Updated Test Environment"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "tags.#", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccTagGroupResource_WithTopic(t *testing.T) {
	// Test tag group with topic and help fields

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTagGroupDestroy,
		Steps: []resource.TestStep{
			// Create tag group with topic
			{
				Config: testAccTagGroupResourceConfigWithTopic(
					"tf_acc_test_location",
					"Test Location",
					"Infrastructure",
					"Location classification for hosts",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "id", "tf_acc_test_location"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "title", "Test Location"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "topic", "Infrastructure"),
					resource.TestCheckResourceAttr("checkmk_tag_group.test", "help", "Location classification for hosts"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_tag_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking"},
			},
		},
	})
}

// testAccCheckTagGroupDestroy verifies that tag groups have been destroyed
func testAccCheckTagGroupDestroy(s *terraform.State) error {
	// Create a client using environment variables
	c, err := client.NewClient(
		testAccTagGroupGetEnv("CHECKMK_URL"),
		testAccTagGroupGetEnv("CHECKMK_USERNAME"),
		testAccTagGroupGetEnv("CHECKMK_PASSWORD"),
	)
	if err != nil {
		return fmt.Errorf("failed to create client for destroy check: %s", err)
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "checkmk_tag_group" {
			continue
		}

		tagGroupID := rs.Primary.ID

		// Try to get the tag group - it should not exist
		_, err := c.GetTagGroup(ctx, tagGroupID)
		if err == nil {
			return fmt.Errorf("tag group still exists: %s", tagGroupID)
		}

		// Verify it's a 404 error (not found)
		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted tag group %s, got: %v", tagGroupID, err)
			}
		}
	}

	return nil
}

// testAccTagGroupGetEnv gets an environment variable, panicking if not set
func testAccTagGroupGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s environment variable must be set for acceptance tests", key))
	}
	return value
}

// testAccTagGroupResourceConfig generates a basic tag group configuration
func testAccTagGroupResourceConfig(id, title string) string {
	return fmt.Sprintf(`
resource "checkmk_tag_group" "test" {
  id    = %[1]q
  title = %[2]q

  tags = [
    {
      id       = "dev"
      title    = "Development"
      aux_tags = []
    },
    {
      id       = "prod"
      title    = "Production"
      aux_tags = []
    },
  ]
}
`, id, title)
}

// testAccTagGroupResourceConfigUpdated generates an updated tag group configuration with more tags
func testAccTagGroupResourceConfigUpdated(id, title string) string {
	return fmt.Sprintf(`
resource "checkmk_tag_group" "test" {
  id    = %[1]q
  title = %[2]q

  tags = [
    {
      id       = "dev"
      title    = "Development"
      aux_tags = []
    },
    {
      id       = "staging"
      title    = "Staging"
      aux_tags = []
    },
    {
      id       = "prod"
      title    = "Production"
      aux_tags = []
    },
  ]
}
`, id, title)
}

// testAccTagGroupResourceConfigWithTopic generates a tag group configuration with topic and help
func testAccTagGroupResourceConfigWithTopic(id, title, topic, help string) string {
	return fmt.Sprintf(`
resource "checkmk_tag_group" "test" {
  id    = %[1]q
  title = %[2]q
  topic = %[3]q
  help  = %[4]q

  tags = [
    {
      id       = "dc1"
      title    = "Data Center 1"
      aux_tags = []
    },
    {
      id       = "dc2"
      title    = "Data Center 2"
      aux_tags = []
    },
  ]
}
`, id, title, topic, help)
}
