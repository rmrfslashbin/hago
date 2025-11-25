package hago

import (
	"context"
	"fmt"
)

// AutomationTriggerRequest contains parameters for triggering an automation.
type AutomationTriggerRequest struct {
	EntityID      string `json:"entity_id"`
	SkipCondition *bool  `json:"skip_condition,omitempty"`
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
