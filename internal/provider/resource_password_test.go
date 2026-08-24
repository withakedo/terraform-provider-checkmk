package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/withakedo/terraform-provider-checkmk/internal/client"
)

func TestAccPasswordResource(t *testing.T) {
	// Passwords do NOT require activation

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPasswordDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPasswordResourceConfig("tf_acc_test_pwd", "Test Password", "secret123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_password.test", "password_id", "tf_acc_test_pwd"),
					resource.TestCheckResourceAttr("checkmk_password.test", "title", "Test Password"),
					resource.TestCheckResourceAttr("checkmk_password.test", "password", "secret123"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "checkmk_password.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strict_resource_locking", "password"},
			},
			// Update and Read testing
			{
				Config: testAccPasswordResourceConfig("tf_acc_test_pwd", "Updated Password", "newsecret456"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_password.test", "password_id", "tf_acc_test_pwd"),
					resource.TestCheckResourceAttr("checkmk_password.test", "title", "Updated Password"),
					resource.TestCheckResourceAttr("checkmk_password.test", "password", "newsecret456"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPasswordResource_WithOptionalFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPasswordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPasswordResourceConfigWithOptionalFields(
					"tf_acc_test_pwd_full",
					"Test Password Full",
					"secret789",
					"admin",
					"https://docs.example.com",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_password.test", "password_id", "tf_acc_test_pwd_full"),
					resource.TestCheckResourceAttr("checkmk_password.test", "title", "Test Password Full"),
					resource.TestCheckResourceAttr("checkmk_password.test", "owner", "admin"),
					resource.TestCheckResourceAttr("checkmk_password.test", "documentation_url", "https://docs.example.com"),
				),
			},
		},
	})
}

func TestAccPasswordResource_WithContactGroups(t *testing.T) {
	// This test uses editable_by with contact groups, which is only supported in CheckMK 2.3+
	// In CheckMK 2.2, the owner field has different semantics and doesn't work with contact groups
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckMinVersion(t, 2, 3) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPasswordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPasswordResourceConfigWithContactGroups(
					"tf_acc_test_pwd_cg",
					"Test Password with Contact Groups",
					"secret999",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkmk_password.test", "password_id", "tf_acc_test_pwd_cg"),
					resource.TestCheckResourceAttr("checkmk_password.test", "title", "Test Password with Contact Groups"),
					resource.TestCheckResourceAttr("checkmk_password.test", "editable_by", "tf_test_editable"),
					resource.TestCheckResourceAttr("checkmk_password.test", "share_with.#", "1"),
				),
			},
		},
	})
}

// testAccCheckPasswordDestroy verifies that passwords have been destroyed
func testAccCheckPasswordDestroy(s *terraform.State) error {
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
		if rs.Type != "checkmk_password" {
			continue
		}

		passwordID := rs.Primary.ID

		_, err := c.GetPassword(ctx, passwordID)
		if err == nil {
			return fmt.Errorf("password still exists: %s", passwordID)
		}

		if apiErr, ok := err.(*client.APIError); ok {
			if apiErr.Status != 404 {
				return fmt.Errorf("expected 404 error for deleted password %s, got: %v", passwordID, err)
			}
		}
	}

	return nil
}

func testAccPasswordResourceConfig(passwordID, title, password string) string {
	return fmt.Sprintf(`
resource "checkmk_password" "test" {
  password_id = %[1]q
  title       = %[2]q
  password    = %[3]q
}
`, passwordID, title, password)
}

func testAccPasswordResourceConfigWithOptionalFields(passwordID, title, password, owner, docURL string) string {
	return fmt.Sprintf(`
resource "checkmk_password" "test" {
  password_id       = %[1]q
  title             = %[2]q
  password          = %[3]q
  owner             = %[4]q
  documentation_url = %[5]q
  comment           = "Test comment"
}
`, passwordID, title, password, owner, docURL)
}

func testAccPasswordResourceConfigWithContactGroups(passwordID, title, password string) string {
	return fmt.Sprintf(`
resource "checkmk_contact_group" "test_editable" {
  name  = "tf_test_editable"
  alias = "Test Editable Group"
}

resource "checkmk_contact_group" "test_shared" {
  name  = "tf_test_shared"
  alias = "Test Shared Group"
}

resource "checkmk_password" "test" {
  password_id = %[1]q
  title       = %[2]q
  password    = %[3]q
  editable_by = checkmk_contact_group.test_editable.name
  share_with  = [checkmk_contact_group.test_shared.name]

  depends_on = [
    checkmk_contact_group.test_editable,
    checkmk_contact_group.test_shared,
  ]
}
`, passwordID, title, password)
}
