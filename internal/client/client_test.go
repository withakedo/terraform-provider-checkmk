package client

import (
	"context"
	"os"
	"testing"
)

func TestClient_GetVersion(t *testing.T) {
	// Skip if not configured
	url := os.Getenv("CHECKMK_URL")
	if url == "" {
		t.Skip("CHECKMK_URL not set")
	}

	username := os.Getenv("CHECKMK_USERNAME")
	password := os.Getenv("CHECKMK_PASSWORD")

	client, err := NewClient(url, username, password)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Logf("Connected to CheckMK version: %s", client.Version.String())

	// Verify version format
	if client.Version.Major == 0 {
		t.Error("Expected non-zero major version")
	}
}

func TestClient_CreateGetDeleteHost(t *testing.T) {
	url := os.Getenv("CHECKMK_URL")
	if url == "" {
		t.Skip("CHECKMK_URL not set")
	}

	username := os.Getenv("CHECKMK_USERNAME")
	password := os.Getenv("CHECKMK_PASSWORD")

	client, err := NewClient(url, username, password)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	testHostName := "test-terraform-host"

	// Cleanup before test
	_ = client.DeleteHost(ctx, testHostName, "")

	// Create host
	t.Run("Create", func(t *testing.T) {
		req := &HostCreateRequest{
			HostName: testHostName,
			Folder:   "/",
			Attributes: map[string]interface{}{
				"alias":     "Test Host for Terraform Provider",
				"ipaddress": "127.0.0.1",
			},
		}

		host, err := client.CreateHost(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}

		if host.ID != testHostName {
			t.Errorf("Expected host ID %s, got %s", testHostName, host.ID)
		}

		t.Logf("Created host: %s", host.ID)
	})

	// Get host
	t.Run("Get", func(t *testing.T) {
		host, err := client.GetHost(ctx, testHostName)
		if err != nil {
			t.Fatalf("Failed to get host: %v", err)
		}

		if host.ID != testHostName {
			t.Errorf("Expected host ID %s, got %s", testHostName, host.ID)
		}

		t.Logf("Retrieved host: %s in folder %s", host.ID, host.Extensions.Folder)
	})

	// Update host
	t.Run("Update", func(t *testing.T) {
		req := &HostUpdateRequest{
			Attributes: map[string]interface{}{
				"alias": "Updated Test Host",
			},
		}

		host, err := client.UpdateHost(ctx, testHostName, req, "")
		if err != nil {
			t.Fatalf("Failed to update host: %v", err)
		}

		t.Logf("Updated host: %s", host.ID)
	})

	// Delete host
	t.Run("Delete", func(t *testing.T) {
		err := client.DeleteHost(ctx, testHostName, "")
		if err != nil {
			t.Fatalf("Failed to delete host: %v", err)
		}

		t.Logf("Deleted host: %s", testHostName)

		// Verify deletion
		_, err = client.GetHost(ctx, testHostName)
		if err == nil {
			t.Error("Expected error when getting deleted host")
		}
	})
}
