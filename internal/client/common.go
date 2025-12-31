package client

import (
	"fmt"
	"net/http"
)

// Link represents a HATEOAS link in the API response
type Link struct {
	DomainType string `json:"domainType,omitempty"`
	Rel        string `json:"rel"`
	Href       string `json:"href"`
	Method     string `json:"method,omitempty"`
	Type       string `json:"type,omitempty"`
}

// BuildETagHeaders creates the If-Match header map for ETag-based operations.
// If etag is non-empty, uses it for strict resource locking.
// If etag is empty, uses "*" to bypass ETag validation.
func BuildETagHeaders(etag string) map[string]string {
	headers := make(map[string]string)
	if etag != "" {
		headers["If-Match"] = etag
	} else {
		headers["If-Match"] = "*"
	}
	return headers
}

// NewNotFoundError creates a standardized 404 Not Found API error
func NewNotFoundError(resourceType, identifier string) *APIError {
	return &APIError{
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("%s '%s' not found", resourceType, identifier),
	}
}

// NewConflictError creates a standardized 409 Conflict API error
func NewConflictError(resourceType, identifier string) *APIError {
	return &APIError{
		Title:  "Conflict",
		Status: http.StatusConflict,
		Detail: fmt.Sprintf("%s '%s' already exists", resourceType, identifier),
	}
}

// NewDriftError creates a standardized drift detection error for ETag mismatches
func NewDriftError(resourceType, identifier string) *DriftError {
	return &DriftError{
		Resource: resourceType,
		Name:     identifier,
		Message:  fmt.Sprintf("%s was modified outside of Terraform (ETag mismatch)", resourceType),
	}
}

// HandleNotFound checks if response is 404 and returns appropriate error
func HandleNotFound(resp *http.Response, resourceType, identifier string) *APIError {
	if resp.StatusCode == http.StatusNotFound {
		return NewNotFoundError(resourceType, identifier)
	}
	return nil
}

// HandleConflict checks if response is 409 and returns appropriate error
func HandleConflict(resp *http.Response, resourceType, identifier string) *APIError {
	if resp.StatusCode == http.StatusConflict {
		return NewConflictError(resourceType, identifier)
	}
	return nil
}

// HandlePreconditionFailed checks if response is 412 and returns appropriate drift error
func HandlePreconditionFailed(resp *http.Response, resourceType, identifier string) *DriftError {
	if resp.StatusCode == http.StatusPreconditionFailed {
		return NewDriftError(resourceType, identifier)
	}
	return nil
}

// IsNotFoundResponse checks if the response indicates a 404 Not Found
func IsNotFoundResponse(resp *http.Response) bool {
	return resp.StatusCode == http.StatusNotFound
}

// IsConflictResponse checks if the response indicates a 409 Conflict
func IsConflictResponse(resp *http.Response) bool {
	return resp.StatusCode == http.StatusConflict
}

// IsPreconditionFailedResponse checks if the response indicates a 412 Precondition Failed
func IsPreconditionFailedResponse(resp *http.Response) bool {
	return resp.StatusCode == http.StatusPreconditionFailed
}

// ListResponse is a generic wrapper for list API responses
type ListResponse[T any] struct {
	Value []T `json:"value"`
}
