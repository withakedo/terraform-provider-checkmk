package client

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
)

// Version represents a CheckMK version
type Version struct {
	Major int
	Minor int
	Patch int
	Build int    // e.g., p41 -> 41
	Raw   string // e.g., "2.2.0p43"
}

// VersionResponse is the API response for /version endpoint
type VersionResponse struct {
	Versions struct {
		Checkmk string `json:"checkmk"`
	} `json:"versions"`
}

// GetVersion retrieves the CheckMK version from the API
func (c *Client) GetVersion(ctx context.Context) (*Version, error) {
	resp, err := c.request(ctx, "GET", "/version", nil)
	if err != nil {
		return nil, err
	}

	var versionResp VersionResponse
	if err := c.handleResponse(resp, &versionResp); err != nil {
		return nil, err
	}

	return ParseVersion(versionResp.Versions.Checkmk)
}

// DetectVersion is an alias for GetVersion for API consistency
func (c *Client) DetectVersion(ctx context.Context) (*Version, error) {
	return c.GetVersion(ctx)
}

// ParseVersion parses a version string like "2.2.0p43" or "2.2.0p43.cee"
func ParseVersion(s string) (*Version, error) {
	// Match version with optional edition suffix (.cee, .cre, .cme, etc.)
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:p(\d+))?(?:\.\w+)?$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", s)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	build := 0
	if matches[4] != "" {
		build, _ = strconv.Atoi(matches[4])
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
		Raw:   s,
	}, nil
}

// AtLeast returns true if this version is >= the given version
// Usage: version.AtLeast(2, 3) or version.AtLeast(2, 2, 0, 5)
func (v *Version) AtLeast(major, minor int, optional ...int) bool {
	patch := 0
	build := 0

	if len(optional) > 0 {
		patch = optional[0]
	}
	if len(optional) > 1 {
		build = optional[1]
	}

	if v.Major != major {
		return v.Major > major
	}
	if v.Minor != minor {
		return v.Minor > minor
	}
	if v.Patch != patch {
		return v.Patch > patch
	}
	return v.Build >= build
}

// GreaterThanOrEqual compares versions (deprecated: use AtLeast)
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}
	return v.Build >= other.Build
}

// String returns the version as a string
func (v *Version) String() string {
	return v.Raw
}
