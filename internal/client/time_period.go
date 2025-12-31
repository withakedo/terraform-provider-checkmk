package client

import (
	"context"
)

// TimePeriod represents a CheckMK time period
type TimePeriod struct {
	ID         string               `json:"id"`
	Title      string               `json:"title"`
	Extensions TimePeriodExtensions `json:"extensions"`
	Links      []Link               `json:"links,omitempty"`
}

// TimePeriodExtensions contains time period extension data
type TimePeriodExtensions struct {
	Alias            string           `json:"alias"`
	ActiveTimeRanges []ActiveTimeDay  `json:"active_time_ranges"`
	Exceptions       []TimePeriodDate `json:"exceptions,omitempty"`
	Exclude          []string         `json:"exclude,omitempty"`
}

// ActiveTimeDay represents time ranges for a specific day
type ActiveTimeDay struct {
	Day        string      `json:"day"`
	TimeRanges []TimeRange `json:"time_ranges"`
}

// TimeRange represents a time range with start and end
type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TimePeriodDate represents an exception date with time ranges
type TimePeriodDate struct {
	Date       string      `json:"date"`
	TimeRanges []TimeRange `json:"time_ranges,omitempty"`
}

// TimePeriodCreateRequest is the request body for creating a time period
type TimePeriodCreateRequest struct {
	Name             string           `json:"name"`
	Alias            string           `json:"alias"`
	ActiveTimeRanges []ActiveTimeDay  `json:"active_time_ranges"`
	Exceptions       []TimePeriodDate `json:"exceptions,omitempty"`
	Exclude          []string         `json:"exclude,omitempty"`
}

// TimePeriodUpdateRequest is the request body for updating a time period
type TimePeriodUpdateRequest struct {
	Alias            string           `json:"alias"`
	ActiveTimeRanges []ActiveTimeDay  `json:"active_time_ranges"`
	Exceptions       []TimePeriodDate `json:"exceptions,omitempty"`
	Exclude          []string         `json:"exclude,omitempty"`
}

// TimePeriodWithETag wraps a TimePeriod with its ETag for strict resource locking
type TimePeriodWithETag struct {
	TimePeriod *TimePeriod
	ETag       string
}

// CreateTimePeriod creates a new time period in CheckMK
func (c *Client) CreateTimePeriod(ctx context.Context, req *TimePeriodCreateRequest) (*TimePeriod, error) {
	return CreateResource[TimePeriod](c, ctx, TimePeriodConfig, req, req.Name)
}

// GetTimePeriod retrieves a time period by name
func (c *Client) GetTimePeriod(ctx context.Context, name string) (*TimePeriod, error) {
	return GetResource[TimePeriod](c, ctx, TimePeriodConfig, name)
}

// GetTimePeriodWithETag retrieves a time period by name and returns it with its ETag
func (c *Client) GetTimePeriodWithETag(ctx context.Context, name string) (*TimePeriodWithETag, error) {
	result, err := GetResourceWithETag[TimePeriod](c, ctx, TimePeriodConfig, name)
	if err != nil {
		return nil, err
	}
	return &TimePeriodWithETag{
		TimePeriod: result.Resource,
		ETag:       result.ETag,
	}, nil
}

// UpdateTimePeriod updates an existing time period
func (c *Client) UpdateTimePeriod(ctx context.Context, name string, req *TimePeriodUpdateRequest, etag string) (*TimePeriod, error) {
	return UpdateResource[TimePeriod](c, ctx, TimePeriodConfig, name, req, etag)
}

// DeleteTimePeriod deletes a time period
func (c *Client) DeleteTimePeriod(ctx context.Context, name string, etag string) error {
	return DeleteResource(c, ctx, TimePeriodConfig, name, etag)
}

// ListTimePeriods retrieves all time periods
func (c *Client) ListTimePeriods(ctx context.Context) ([]TimePeriod, error) {
	return ListResources[TimePeriod](c, ctx, TimePeriodConfig)
}
