package client

import (
	"context"
)

// AcknowledgeHostProblemRequest acknowledges the current problem state of a
// single host. Mirrors the "AcknowledgeHostProblem" schema (acknowledge_type
// = "host") from the CheckMK REST API.
type AcknowledgeHostProblemRequest struct {
	AcknowledgeType string `json:"acknowledge_type"`
	HostName        string `json:"host_name"`
	Comment         string `json:"comment"`
	Sticky          bool   `json:"sticky"`
	Persistent      bool   `json:"persistent"`
	Notify          bool   `json:"notify"`
	ExpireOn        string `json:"expire_on,omitempty"`
}

// AcknowledgeServiceProblemRequest acknowledges the current problem state of
// a single service on a host. Mirrors the "AcknowledgeSpecificServiceProblem"
// schema (acknowledge_type = "service") from the CheckMK REST API.
type AcknowledgeServiceProblemRequest struct {
	AcknowledgeType    string `json:"acknowledge_type"`
	HostName           string `json:"host_name"`
	ServiceDescription string `json:"service_description"`
	Comment            string `json:"comment"`
	Sticky             bool   `json:"sticky"`
	Persistent         bool   `json:"persistent"`
	Notify             bool   `json:"notify"`
	ExpireOn           string `json:"expire_on,omitempty"`
}

// RemoveAcknowledgementRequest removes the acknowledgement on a host or
// service problem. Mirrors the "RemoveProblemAcknowledgement" discriminated
// union (ViaSpecificHost / ViaSpecificService variants) from the CheckMK
// REST API.
type RemoveAcknowledgementRequest struct {
	AcknowledgeType    string `json:"acknowledge_type"`
	HostName           string `json:"host_name"`
	ServiceDescription string `json:"service_description,omitempty"`
}

// AcknowledgeHostProblem acknowledges the current problem state of a host.
func (c *Client) AcknowledgeHostProblem(ctx context.Context, req *AcknowledgeHostProblemRequest) error {
	resp, err := c.request(ctx, "POST", "/domain-types/acknowledge/collections/host", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", req.HostName); apiErr != nil {
		resp.Body.Close()
		return apiErr
	}
	return c.handleResponse(resp, nil)
}

// AcknowledgeServiceProblem acknowledges the current problem state of a
// service on a host.
func (c *Client) AcknowledgeServiceProblem(ctx context.Context, req *AcknowledgeServiceProblemRequest) error {
	resp, err := c.request(ctx, "POST", "/domain-types/acknowledge/collections/service", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", req.HostName); apiErr != nil {
		resp.Body.Close()
		return apiErr
	}
	return c.handleResponse(resp, nil)
}

// RemoveAcknowledgement removes the acknowledgement on a host, or on a
// specific service on a host if serviceDescription is non-empty. Removing an
// acknowledgement that no longer exists (e.g. the problem already resolved
// naturally) is treated as success, matching the idempotent-delete
// convention used elsewhere in this client.
func (c *Client) RemoveAcknowledgement(ctx context.Context, hostName, serviceDescription string) error {
	req := &RemoveAcknowledgementRequest{
		AcknowledgeType:    "host",
		HostName:           hostName,
		ServiceDescription: serviceDescription,
	}
	if serviceDescription != "" {
		req.AcknowledgeType = "service"
	}

	resp, err := c.request(ctx, "POST", "/domain-types/acknowledge/actions/delete/invoke", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", hostName); apiErr != nil {
		resp.Body.Close()
		return nil
	}
	return c.handleResponse(resp, nil)
}
