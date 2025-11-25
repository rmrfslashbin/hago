package hago

import (
	"context"
	"fmt"
	"net/http"
)

// AutomationTriggerRequest contains parameters for triggering an automation.
type AutomationTriggerRequest struct {
	EntityID      string `json:"entity_id"`
	SkipCondition *bool  `json:"skip_condition,omitempty"`
}

// AutomationConfig represents a complete automation configuration.
//
// WARNING: This uses an undocumented Home Assistant REST API endpoint that is
// subject to change without notice. The endpoint /api/config/automation/config/
// is used internally by the Home Assistant UI but is not officially documented.
// Use at your own risk and expect potential breaking changes in future HA versions.
type AutomationConfig struct {
	ID          string         `json:"id"`
	Alias       string         `json:"alias"`
	Description *string        `json:"description,omitempty"`
	Mode        string         `json:"mode,omitempty"`        // single, restart, parallel, queued
	MaxExceeded *string        `json:"max_exceeded,omitempty"` // warn, silent
	Max         *int           `json:"max,omitempty"`
	Trigger     []any          `json:"trigger"`
	Condition   []any          `json:"condition,omitempty"`
	Action      []any          `json:"action"`
}

// AutomationTrigger triggers an automation, optionally skipping conditions.
func (c *Client) AutomationTrigger(ctx context.Context, req *AutomationTriggerRequest) error {
	if req == nil || req.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	serviceReq := &ServiceCallRequest{
		EntityID: req.EntityID,
	}

	// Only include skip_condition if it's set
	if req.SkipCondition != nil {
		serviceReq.Data = map[string]any{
			"skip_condition": *req.SkipCondition,
		}
	}

	_, err := c.CallService(ctx, "automation", "trigger", serviceReq)
	if err != nil {
		return fmt.Errorf("automation trigger: %w", err)
	}
	return nil
}

// AutomationTurnOn enables an automation.
func (c *Client) AutomationTurnOn(ctx context.Context, entityID string) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	_, err := c.CallService(ctx, "automation", "turn_on", &ServiceCallRequest{
		EntityID: entityID,
	})
	if err != nil {
		return fmt.Errorf("automation turn_on: %w", err)
	}
	return nil
}

// AutomationTurnOff disables an automation.
func (c *Client) AutomationTurnOff(ctx context.Context, entityID string) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	_, err := c.CallService(ctx, "automation", "turn_off", &ServiceCallRequest{
		EntityID: entityID,
	})
	if err != nil {
		return fmt.Errorf("automation turn_off: %w", err)
	}
	return nil
}

// AutomationToggle toggles an automation's enabled state.
func (c *Client) AutomationToggle(ctx context.Context, entityID string) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	_, err := c.CallService(ctx, "automation", "toggle", &ServiceCallRequest{
		EntityID: entityID,
	})
	if err != nil {
		return fmt.Errorf("automation toggle: %w", err)
	}
	return nil
}

// AutomationReload reloads all automations from YAML configuration.
func (c *Client) AutomationReload(ctx context.Context) error {
	_, err := c.CallService(ctx, "automation", "reload", nil)
	if err != nil {
		return fmt.Errorf("automation reload: %w", err)
	}
	return nil
}

// AutomationList lists all automation configurations.
//
// WARNING: This uses an undocumented REST API endpoint (/api/config/automation/config)
// that is subject to change. See AutomationConfig for details.
func (c *Client) AutomationList(ctx context.Context) ([]AutomationConfig, error) {
	var configs []AutomationConfig
	if err := c.doJSON(ctx, http.MethodGet, "/api/config/automation/config", nil, &configs); err != nil {
		return nil, fmt.Errorf("automation list: %w", err)
	}
	return configs, nil
}

// AutomationGet retrieves a specific automation configuration by ID.
//
// WARNING: This uses an undocumented REST API endpoint. See AutomationConfig for details.
func (c *Client) AutomationGet(ctx context.Context, id string) (*AutomationConfig, error) {
	if id == "" {
		return nil, fmt.Errorf("automation id is required")
	}

	var config AutomationConfig
	path := fmt.Sprintf("/api/config/automation/config/%s", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &config); err != nil {
		return nil, fmt.Errorf("automation get: %w", err)
	}
	return &config, nil
}

// AutomationSave creates or updates an automation configuration.
//
// WARNING: This uses an undocumented REST API endpoint. See AutomationConfig for details.
func (c *Client) AutomationSave(ctx context.Context, config *AutomationConfig) error {
	if config == nil {
		return fmt.Errorf("automation config is required")
	}
	if config.ID == "" {
		return fmt.Errorf("automation id is required")
	}
	if config.Alias == "" {
		return fmt.Errorf("automation alias is required")
	}

	path := fmt.Sprintf("/api/config/automation/config/%s", config.ID)
	if err := c.doJSON(ctx, http.MethodPost, path, config, nil); err != nil {
		return fmt.Errorf("automation save: %w", err)
	}
	return nil
}

// AutomationDeleteConfig deletes an automation configuration by ID.
//
// WARNING: This uses an undocumented REST API endpoint. See AutomationConfig for details.
func (c *Client) AutomationDeleteConfig(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("automation id is required")
	}

	path := fmt.Sprintf("/api/config/automation/config/%s", id)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("automation delete: %w", err)
	}
	return nil
}
