package client

import (
	"context"
	"net/http"
)

// CreateHostDowntimeRequest schedules a downtime for a single host.
// Mirrors the "CreateHostDowntime" schema (downtime_type = "host") from the
// CheckMK REST API.
type CreateHostDowntimeRequest struct {
	DowntimeType string `json:"downtime_type"`
	HostName     string `json:"host_name"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Comment      string `json:"comment,omitempty"`
	Duration     int64  `json:"duration,omitempty"`
	Recur        string `json:"recur,omitempty"`
}

// CreateServiceDowntimeRequest schedules a downtime for one or more services
// on a host. Mirrors the "CreateServiceDowntime" schema (downtime_type =
// "service") from the CheckMK REST API.
type CreateServiceDowntimeRequest struct {
	DowntimeType        string   `json:"downtime_type"`
	HostName            string   `json:"host_name"`
	ServiceDescriptions []string `json:"service_descriptions"`
	StartTime           string   `json:"start_time"`
	EndTime             string   `json:"end_time"`
	Comment             string   `json:"comment,omitempty"`
	Duration            int64    `json:"duration,omitempty"`
	Recur               string   `json:"recur,omitempty"`
}

// DeleteDowntimeByParamsRequest deletes downtime(s) by host (and optionally
// specific services) rather than by server-generated downtime ID. CheckMK's
// create-downtime endpoints return 204 No Content with no id, so matching by
// the original parameters avoids having to look up the generated id/site_id
// pair afterward.
type DeleteDowntimeByParamsRequest struct {
	DeleteType          string   `json:"delete_type"`
	HostName            string   `json:"host_name"`
	ServiceDescriptions []string `json:"service_descriptions,omitempty"`
}

// CreateHostDowntime schedules a downtime for a host.
func (c *Client) CreateHostDowntime(ctx context.Context, req *CreateHostDowntimeRequest) error {
	resp, err := c.request(ctx, "POST", "/domain-types/downtime/collections/host", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", req.HostName); apiErr != nil {
		resp.Body.Close()
		return apiErr
	}
	return c.handleResponse(resp, nil)
}

// CreateServiceDowntime schedules a downtime for one or more services on a host.
func (c *Client) CreateServiceDowntime(ctx context.Context, req *CreateServiceDowntimeRequest) error {
	resp, err := c.request(ctx, "POST", "/domain-types/downtime/collections/service", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", req.HostName); apiErr != nil {
		resp.Body.Close()
		return apiErr
	}
	return c.handleResponse(resp, nil)
}

// DeleteDowntimeByParams removes the downtime(s) matching a host and,
// optionally, a specific set of services. Deleting a downtime that no longer
// exists (e.g. it already expired naturally) is treated as success, matching
// the idempotent-delete convention used elsewhere in this client.
func (c *Client) DeleteDowntimeByParams(ctx context.Context, hostName string, serviceDescriptions []string) error {
	req := &DeleteDowntimeByParamsRequest{
		DeleteType:          "params",
		HostName:            hostName,
		ServiceDescriptions: serviceDescriptions,
	}

	resp, err := c.request(ctx, "POST", "/domain-types/downtime/actions/delete/invoke", req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil
	}
	return c.handleResponse(resp, nil)
}
