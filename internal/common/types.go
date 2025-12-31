package common

import (
	"sync"

	"github.com/terraform-provider-checkmk/internal/client"
)

// ProviderData holds the configured provider data including client and settings.
// This is passed to all resources during configuration.
type ProviderData struct {
	Client                *client.Client
	Activate              string
	ForceForeignChanges   bool
	ActivationWaitTime    int
	StrictResourceLocking bool
	Features              *Features

	// PendingChanges tracks changes made by resources for batch activation reporting.
	// Key is resource type (e.g., "folder", "host", "rule"), value is count.
	PendingChanges   map[string]int
	pendingChangesMu sync.Mutex
}

// TrackChange increments the pending change counter for a resource type.
// This is called by resources when they make changes in manual activation mode.
func (p *ProviderData) TrackChange(resourceType string) {
	p.pendingChangesMu.Lock()
	defer p.pendingChangesMu.Unlock()
	if p.PendingChanges == nil {
		p.PendingChanges = make(map[string]int)
	}
	p.PendingChanges[resourceType]++
}

// GetPendingChanges returns a copy of the pending changes map.
func (p *ProviderData) GetPendingChanges() map[string]int {
	p.pendingChangesMu.Lock()
	defer p.pendingChangesMu.Unlock()
	if p.PendingChanges == nil {
		return make(map[string]int)
	}
	// Return a copy to avoid race conditions
	result := make(map[string]int, len(p.PendingChanges))
	for k, v := range p.PendingChanges {
		result[k] = v
	}
	return result
}

// ClearPendingChanges resets the pending changes counter.
func (p *ProviderData) ClearPendingChanges() {
	p.pendingChangesMu.Lock()
	defer p.pendingChangesMu.Unlock()
	p.PendingChanges = make(map[string]int)
}

// TotalPendingChanges returns the total count of pending changes across all resource types.
func (p *ProviderData) TotalPendingChanges() int {
	p.pendingChangesMu.Lock()
	defer p.pendingChangesMu.Unlock()
	total := 0
	for _, count := range p.PendingChanges {
		total += count
	}
	return total
}

// BaseResourceConfig holds common configuration for all resources
type BaseResourceConfig struct {
	Client                *client.Client
	Activate              string
	ForceForeignChanges   bool
	ActivationWaitTime    int
	StrictResourceLocking bool
}
