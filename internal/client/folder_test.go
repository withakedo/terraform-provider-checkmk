package client

import (
	"testing"
)

func TestPathToAPIFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "root folder",
			input:    "/",
			expected: "~",
		},
		{
			name:     "first level folder",
			input:    "/production",
			expected: "~production",
		},
		{
			name:     "nested folder",
			input:    "/IE/MUL/Server",
			expected: "~IE~MUL~Server",
		},
		{
			name:     "deep nested folder",
			input:    "/country/location/type/subtype",
			expected: "~country~location~type~subtype",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathToAPIFormat(tt.input)
			if result != tt.expected {
				t.Errorf("PathToAPIFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPathFromAPIFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "root folder",
			input:    "~",
			expected: "/",
		},
		{
			name:     "first level folder",
			input:    "~production",
			expected: "/production",
		},
		{
			name:     "nested folder",
			input:    "~IE~MUL~Server",
			expected: "/IE/MUL/Server",
		},
		{
			name:     "deep nested folder",
			input:    "~country~location~type~subtype",
			expected: "/country/location/type/subtype",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathFromAPIFormat(tt.input)
			if result != tt.expected {
				t.Errorf("PathFromAPIFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPathRoundTrip(t *testing.T) {
	// Test that converting to API format and back gives the original path
	paths := []string{
		"/",
		"/production",
		"/IE/MUL",
		"/IE/MUL/Server",
		"/country/location/type/subtype",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			apiFormat := PathToAPIFormat(path)
			result := PathFromAPIFormat(apiFormat)
			if result != path {
				t.Errorf("Round trip failed: %q -> %q -> %q", path, apiFormat, result)
			}
		})
	}
}
