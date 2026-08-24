package client

import (
	"context"
	"fmt"
	"net/http"
)

// DiscoverServicesRequest triggers a service discovery run for a host.
// Mirrors the "DiscoverServices" schema from the CheckMK REST API.
type DiscoverServicesRequest struct {
	HostName string `json:"host_name"`
	Mode     string `json:"mode,omitempty"`
}

// ServiceDiscoveryRunResponse is the response from starting or polling a
// discovery run (GET /objects/service_discovery_run/{host_name}).
type ServiceDiscoveryRunResponse struct {
	ID         string                        `json:"id"`
	DomainType string                        `json:"domainType"`
	Title      string                        `json:"title"`
	Extensions ServiceDiscoveryRunExtensions `json:"extensions"`
}

// ServiceDiscoveryRunExtensions describes the state of a discovery run.
type ServiceDiscoveryRunExtensions struct {
	Active bool                    `json:"active"`
	State  string                  `json:"state"` // initialized, running, finished, stopped, exception
	Logs   ServiceDiscoveryRunLogs `json:"logs"`
}

// ServiceDiscoveryRunLogs holds progress/result log lines for a discovery run.
type ServiceDiscoveryRunLogs struct {
	Progress []string `json:"progress"`
	Result   []string `json:"result"`
}

// DiscoverServices starts a service discovery run for a host and blocks until
// it finishes, returning the run's final state. mode controls how discovered
// services are handled server-side (e.g. "fix_all" and "tabula_rasa" apply
// changes immediately; "refresh" only updates the discovery preview without
// changing the host's active service configuration).
func (c *Client) DiscoverServices(ctx context.Context, hostName, mode string) (*ServiceDiscoveryRunExtensions, error) {
	req := &DiscoverServicesRequest{
		HostName: hostName,
		Mode:     mode,
	}

	resp, err := c.request(ctx, "POST", "/domain-types/service_discovery_run/actions/start/invoke", req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusConflict {
		resp.Body.Close()
		return nil, NewConflictError("Service discovery run", hostName)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, NewNotFoundError("Host", hostName)
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return nil, err
	}

	return c.waitForDiscoveryCompletion(ctx, hostName)
}

// waitForDiscoveryCompletion polls CheckMK's wait-for-completion endpoint -
// which, like activation's, redirects to itself while the discovery run is
// still in progress - and then fetches the final run status. Polling uses
// LongOperationTimeout rather than the general request_timeout, since
// discovery on a host with many services can legitimately run far longer
// than a normal CRUD call.
func (c *Client) waitForDiscoveryCompletion(ctx context.Context, hostName string) (*ServiceDiscoveryRunExtensions, error) {
	waitURL := fmt.Sprintf("%s/check_mk/api/1.0/objects/service_discovery_run/%s/actions/wait-for-completion/invoke", c.BaseURL.String(), hostName)
	if err := c.pollSelfRedirectingCompletion(ctx, waitURL); err != nil {
		return nil, err
	}

	statusPath := fmt.Sprintf("/objects/service_discovery_run/%s", hostName)
	statusResp, err := c.request(ctx, "GET", statusPath, nil)
	if err != nil {
		return nil, err
	}

	var run ServiceDiscoveryRunResponse
	if err := c.handleResponse(statusResp, &run); err != nil {
		return nil, err
	}

	return &run.Extensions, nil
}
