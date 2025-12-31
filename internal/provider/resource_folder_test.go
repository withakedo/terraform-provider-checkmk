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

func TestAccFolderResource(t *testing.T) {
	// NOTE: This test requires manual activation of changes in CheckMK.
	// After each step, changes must be activated for them to take effect.
	// Future enhancement: Add automatic activation support.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFolderResourceConfig("tf-acc-test-folder", "Test Folder"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_folder.test", "name", "tf-acc-test-folder"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "parent", "/"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "title", "Test Folder"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "path", "/tf-acc-test-folder"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "id", "/tf-acc-test-folder"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "checkmk_folder.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccFolderResourceConfig("tf-acc-test-folder", "Updated Test Folder"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_folder.test", "name", "tf-acc-test-folder"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "title", "Updated Test Folder"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccFolderResource_WithAttributes(t *testing.T) {
	// Test folder creation and updates with attributes

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy,
		Steps: []resource.TestStep{
			// Create folder with attributes
			{
				Config: testAccFolderResourceConfigWithAttributes(
					"tf-acc-test-folder-attrs",
					"Folder with Attributes",
					map[string]string{
						"tag_criticality": "prod",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_folder.test", "name", "tf-acc-test-folder-attrs"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "title", "Folder with Attributes"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "attributes.tag_criticality", "prod"),
				),
			},
			// Update attributes
			{
				Config: testAccFolderResourceConfigWithAttributes(
					"tf-acc-test-folder-attrs",
					"Folder with Attributes",
					map[string]string{
						"tag_criticality": "test",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_folder.test", "name", "tf-acc-test-folder-attrs"),
					resource.TestCheckResourceAttr("checkmk_folder.test", "attributes.tag_criticality", "test"),
				),
			},
			// ImportState testing with attributes
			{
				ResourceName:      "checkmk_folder.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFolderResource_Nested(t *testing.T) {
	// Test nested folder hierarchy

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy,
		Steps: []resource.TestStep{
			// Create parent and child folders
			{
				Config: testAccFolderResourceConfigNested(
					"tf-acc-test-parent",
					"Parent Folder",
					"tf-acc-test-child",
					"Child Folder",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check parent folder
					resource.TestCheckResourceAttr("checkmk_folder.parent", "name", "tf-acc-test-parent"),
					resource.TestCheckResourceAttr("checkmk_folder.parent", "parent", "/"),
					resource.TestCheckResourceAttr("checkmk_folder.parent", "path", "/tf-acc-test-parent"),
					resource.TestCheckResourceAttr("checkmk_folder.parent", "title", "Parent Folder"),
					// Check child folder
					resource.TestCheckResourceAttr("checkmk_folder.child", "name", "tf-acc-test-child"),
					resource.TestCheckResourceAttr("checkmk_folder.child", "parent", "/tf-acc-test-parent"),
					resource.TestCheckResourceAttr("checkmk_folder.child", "path", "/tf-acc-test-parent/tf-acc-test-child"),
					resource.TestCheckResourceAttr("checkmk_folder.child", "title", "Child Folder"),
				),
			},
			// ImportState testing for parent
			{
				ResourceName:      "checkmk_folder.parent",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ImportState testing for child
			{
				ResourceName:      "checkmk_folder.child",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update nested folders
			{
				Config: testAccFolderResourceConfigNested(
					"tf-acc-test-parent",
					"Updated Parent Folder",
					"tf-acc-test-child",
					"Updated Child Folder",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_folder.parent", "title", "Updated Parent Folder"),
					resource.TestCheckResourceAttr("checkmk_folder.child", "title", "Updated Child Folder"),
				),
			},
			// Delete testing automatically occurs in TestCase
			// Terraform will handle deletion order (child before parent)
		},
	})
}

// testAccCheckFolderDestroy verifies that folders have been destroyed
func testAccCheckFolderDestroy(s *terraform.State) error {
	// Create a client using environment variables
	c, err := client.NewClient(
		testAccGetEnv("CHECKMK_URL"),
		testAccGetEnv("CHECKMK_USERNAME"),
		testAccGetEnv("CHECKMK_PASSWORD"),
	)
	if err != nil {
		return fmt.Errorf("failed to create client for destroy check: %s", err)
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "checkmk_folder" {
			continue
		}

		folderPath := rs.Primary.ID

		// Try to get the folder - it should not exist
		_, err := c.GetFolder(ctx, folderPath)
		if err == nil {
			return fmt.Errorf("folder still exists: %s", folderPath)
		}

		// Verify it's a 404 error (not found)
		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted folder %s, got: %v", folderPath, err)
			}
		}
	}

	return nil
}

// testAccGetEnv gets an environment variable, panicking if not set
func testAccGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s environment variable must be set for acceptance tests", key))
	}
	return value
}

// testAccFolderResourceConfig generates a basic folder configuration
func testAccFolderResourceConfig(name, title string) string {
	return fmt.Sprintf(`
resource "checkmk_folder" "test" {
  name   = %[1]q
  parent = "/"
  title  = %[2]q
}
`, name, title)
}

// testAccFolderResourceConfigWithAttributes generates a folder configuration with attributes
func testAccFolderResourceConfigWithAttributes(name, title string, attributes map[string]string) string {
	config := fmt.Sprintf(`
resource "checkmk_folder" "test" {
  name   = %[1]q
  parent = "/"
  title  = %[2]q

  attributes = {
`, name, title)

	for k, v := range attributes {
		config += fmt.Sprintf("    %s = %q\n", k, v)
	}

	config += "  }\n}\n"
	return config
}

// testAccFolderResourceConfigNested generates a configuration with nested folders
func testAccFolderResourceConfigNested(parentName, parentTitle, childName, childTitle string) string {
	return fmt.Sprintf(`
resource "checkmk_folder" "parent" {
  name   = %[1]q
  parent = "/"
  title  = %[2]q
}

resource "checkmk_folder" "child" {
  name   = %[3]q
  parent = checkmk_folder.parent.path
  title  = %[4]q
}
`, parentName, parentTitle, childName, childTitle)
}
