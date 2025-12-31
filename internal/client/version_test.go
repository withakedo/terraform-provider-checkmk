package client

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMajor   int
		wantMinor   int
		wantPatch   int
		wantBuild   int
		wantErr     bool
		description string
	}{
		{
			name:        "Standard version with build",
			input:       "2.2.0p43",
			wantMajor:   2,
			wantMinor:   2,
			wantPatch:   0,
			wantBuild:   43,
			wantErr:     false,
			description: "Typical CheckMK version format",
		},
		{
			name:        "Version with edition suffix",
			input:       "2.3.0p41.cee",
			wantMajor:   2,
			wantMinor:   3,
			wantPatch:   0,
			wantBuild:   41,
			wantErr:     false,
			description: "CheckMK Enterprise Edition version",
		},
		{
			name:        "Version without build number",
			input:       "2.1.0",
			wantMajor:   2,
			wantMinor:   1,
			wantPatch:   0,
			wantBuild:   0,
			wantErr:     false,
			description: "Version without patch level",
		},
		{
			name:        "Latest 2.4.0 version",
			input:       "2.4.0p17",
			wantMajor:   2,
			wantMinor:   4,
			wantPatch:   0,
			wantBuild:   17,
			wantErr:     false,
			description: "CheckMK 2.4.0p17 version",
		},
		{
			name:        "Version with CRE edition",
			input:       "2.2.0p5.cre",
			wantMajor:   2,
			wantMinor:   2,
			wantPatch:   0,
			wantBuild:   5,
			wantErr:     false,
			description: "CheckMK Raw Edition version",
		},
		{
			name:        "Version with CME edition",
			input:       "2.3.0p10.cme",
			wantMajor:   2,
			wantMinor:   3,
			wantPatch:   0,
			wantBuild:   10,
			wantErr:     false,
			description: "CheckMK Managed Services Edition version",
		},
		{
			name:        "Invalid format - missing patch",
			input:       "2.2",
			wantErr:     true,
			description: "Should fail on incomplete version",
		},
		{
			name:        "Invalid format - random string",
			input:       "not-a-version",
			wantErr:     true,
			description: "Should fail on non-version string",
		},
		{
			name:        "Invalid format - extra parts",
			input:       "2.2.0.0.0",
			wantErr:     true,
			description: "Should fail on too many version parts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseVersion(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got.Major != tt.wantMajor {
				t.Errorf("ParseVersion(%q).Major = %d, want %d", tt.input, got.Major, tt.wantMajor)
			}
			if got.Minor != tt.wantMinor {
				t.Errorf("ParseVersion(%q).Minor = %d, want %d", tt.input, got.Minor, tt.wantMinor)
			}
			if got.Patch != tt.wantPatch {
				t.Errorf("ParseVersion(%q).Patch = %d, want %d", tt.input, got.Patch, tt.wantPatch)
			}
			if got.Build != tt.wantBuild {
				t.Errorf("ParseVersion(%q).Build = %d, want %d", tt.input, got.Build, tt.wantBuild)
			}
			if got.Raw != tt.input {
				t.Errorf("ParseVersion(%q).Raw = %q, want %q", tt.input, got.Raw, tt.input)
			}
		})
	}
}

func TestVersion_AtLeast(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		major    int
		minor    int
		optional []int
		want     bool
	}{
		{
			name:    "2.3.0p41 >= 2.3",
			version: &Version{Major: 2, Minor: 3, Patch: 0, Build: 41},
			major:   2,
			minor:   3,
			want:    true,
		},
		{
			name:    "2.3.0p41 >= 2.2",
			version: &Version{Major: 2, Minor: 3, Patch: 0, Build: 41},
			major:   2,
			minor:   2,
			want:    true,
		},
		{
			name:    "2.2.0p43 >= 2.3",
			version: &Version{Major: 2, Minor: 2, Patch: 0, Build: 43},
			major:   2,
			minor:   3,
			want:    false,
		},
		{
			name:     "2.2.0p43 >= 2.2.0p5",
			version:  &Version{Major: 2, Minor: 2, Patch: 0, Build: 43},
			major:    2,
			minor:    2,
			optional: []int{0, 5},
			want:     true,
		},
		{
			name:     "2.2.0p5 >= 2.2.0p5 (equal)",
			version:  &Version{Major: 2, Minor: 2, Patch: 0, Build: 5},
			major:    2,
			minor:    2,
			optional: []int{0, 5},
			want:     true,
		},
		{
			name:     "2.2.0p4 >= 2.2.0p5",
			version:  &Version{Major: 2, Minor: 2, Patch: 0, Build: 4},
			major:    2,
			minor:    2,
			optional: []int{0, 5},
			want:     false,
		},
		{
			name:     "2.4.0p17 >= 2.3.0",
			version:  &Version{Major: 2, Minor: 4, Patch: 0, Build: 17},
			major:    2,
			minor:    3,
			optional: []int{0},
			want:     true,
		},
		{
			name:    "2.1.0 >= 2.2",
			version: &Version{Major: 2, Minor: 1, Patch: 0, Build: 0},
			major:   2,
			minor:   2,
			want:    false,
		},
		{
			name:    "3.0.0 >= 2.3",
			version: &Version{Major: 3, Minor: 0, Patch: 0, Build: 0},
			major:   2,
			minor:   3,
			want:    true,
		},
		{
			name:     "2.2.1 >= 2.2.0",
			version:  &Version{Major: 2, Minor: 2, Patch: 1, Build: 0},
			major:    2,
			minor:    2,
			optional: []int{0},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.AtLeast(tt.major, tt.minor, tt.optional...)
			if got != tt.want {
				t.Errorf("Version.AtLeast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_GreaterThanOrEqual(t *testing.T) {
	v1 := &Version{Major: 2, Minor: 3, Patch: 0, Build: 41}
	v2 := &Version{Major: 2, Minor: 2, Patch: 0, Build: 43}
	v3 := &Version{Major: 2, Minor: 3, Patch: 0, Build: 41}

	if !v1.GreaterThanOrEqual(v2) {
		t.Error("2.3.0p41 should be >= 2.2.0p43")
	}

	if !v1.GreaterThanOrEqual(v3) {
		t.Error("2.3.0p41 should be >= 2.3.0p41 (equal)")
	}

	if v2.GreaterThanOrEqual(v1) {
		t.Error("2.2.0p43 should not be >= 2.3.0p41")
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Standard version", "2.2.0p43"},
		{"With edition", "2.3.0p41.cee"},
		{"Without build", "2.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if err != nil {
				t.Fatalf("ParseVersion(%q) error: %v", tt.input, err)
			}

			if v.String() != tt.input {
				t.Errorf("Version.String() = %q, want %q", v.String(), tt.input)
			}
		})
	}
}
