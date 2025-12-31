package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/terraform-provider-checkmk/internal/client"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"checkmk": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Ensure required environment variables are set
	if os.Getenv("CHECKMK_URL") == "" {
		t.Fatal("CHECKMK_URL must be set for acceptance tests")
	}
	if os.Getenv("CHECKMK_USERNAME") == "" {
		t.Fatal("CHECKMK_USERNAME must be set for acceptance tests")
	}
	if os.Getenv("CHECKMK_PASSWORD") == "" {
		t.Fatal("CHECKMK_PASSWORD must be set for acceptance tests")
	}
}

// testAccPreCheckMinVersion skips the test if the CheckMK version is below the minimum required
func testAccPreCheckMinVersion(t *testing.T, major, minor int) {
	testAccPreCheck(t)

	c, err := client.NewClient(
		os.Getenv("CHECKMK_URL"),
		os.Getenv("CHECKMK_USERNAME"),
		os.Getenv("CHECKMK_PASSWORD"),
	)
	if err != nil {
		t.Fatalf("Failed to create client for version check: %s", err)
	}

	if !c.Version.AtLeast(major, minor) {
		t.Skipf("Test requires CheckMK %d.%d+, current version is %s", major, minor, c.Version.String())
	}
}
