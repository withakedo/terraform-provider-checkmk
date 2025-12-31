package client

import "fmt"

// APIError represents a CheckMK API error response
type APIError struct {
	Title  string              `json:"title"`
	Status int                 `json:"status"`
	Detail string              `json:"detail"`
	Fields map[string][]string `json:"fields,omitempty"`
}

func (e *APIError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf("%s (status %d): %s - Fields: %v", e.Title, e.Status, e.Detail, e.Fields)
	}
	return fmt.Sprintf("%s (status %d): %s", e.Title, e.Status, e.Detail)
}

// DriftError indicates configuration drift (resource modified outside Terraform)
type DriftError struct {
	Resource string
	Name     string
	Message  string
}

func (e *DriftError) Error() string {
	return fmt.Sprintf("configuration drift detected for %s '%s': %s", e.Resource, e.Name, e.Message)
}

// IsDriftError checks if an error is a drift error (HTTP 428)
func IsDriftError(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Status == 428
}
