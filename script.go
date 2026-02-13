package hago

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ScriptConfig represents a complete script configuration.
//
// WARNING: This uses an undocumented Home Assistant REST API endpoint that is
// subject to change without notice. The endpoint /api/config/script/config/
// is used internally by the Home Assistant UI but is not officially documented.
// Use at your own risk and expect potential breaking changes in future HA versions.
type ScriptConfig struct {
	ID          string         `json:"id"`
	Alias       string         `json:"alias"`
	Sequence    []any          `json:"sequence"`
	Mode        string         `json:"mode,omitempty"`        // single, restart, parallel, queued
	Max         *int           `json:"max,omitempty"`
	Icon        *string        `json:"icon,omitempty"`
	Description *string        `json:"description,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
	Variables   map[string]any `json:"variables,omitempty"`
}

// ScriptList lists all script configurations.
//
// WARNING: This uses an undocumented REST API endpoint (/api/config/script/config)
// that is subject to change. See ScriptConfig for details.
//
// Implementation notes:
// - First attempts the undocumented /api/config/script/config endpoint
// - If that fails (common 404 error), falls back to listing script entities via /api/states
// - Fallback mode returns only basic metadata (ID, Alias) without sequence
// - Only returns UI-created scripts stored in scripts.yaml, not YAML-defined ones
func (c *Client) ScriptList(ctx context.Context) ([]ScriptConfig, error) {
	// Try the undocumented config endpoint first
	var configs []ScriptConfig
	err := c.doJSON(ctx, http.MethodGet, "/api/config/script/config", nil, &configs)
	if err == nil {
		return configs, nil
	}

	// If config endpoint fails (common 404), fall back to States API
	// This is documented and reliable but returns less detailed information
	states, statesErr := c.States(ctx)
	if statesErr != nil {
		// Return original error if both methods fail
		return nil, fmt.Errorf("script list: config endpoint failed (%w), states fallback also failed (%v)", err, statesErr)
	}

	// Filter for script.* entities and build basic configs
	configs = make([]ScriptConfig, 0)
	for _, state := range states {
		if !strings.HasPrefix(state.EntityID, "script.") {
			continue
		}

		config := ScriptConfig{
			ID: state.EntityID,
		}

		// Extract friendly_name as alias if available
		if friendlyName, ok := state.Attributes["friendly_name"].(string); ok {
			config.Alias = friendlyName
		} else {
			// Fallback: use entity_id without "script." prefix
			config.Alias = strings.TrimPrefix(state.EntityID, "script.")
		}

		// Extract other attributes if available
		if desc, ok := state.Attributes["description"].(string); ok && desc != "" {
			config.Description = &desc
		}
		if mode, ok := state.Attributes["mode"].(string); ok && mode != "" {
			config.Mode = mode
		}

		// Note: Sequence will be empty in fallback mode
		// Users need to call ScriptGet(id) for full configuration
		config.Sequence = []any{}

		configs = append(configs, config)
	}

	return configs, nil
}

// ScriptGet retrieves a specific script configuration by ID.
//
// WARNING: This uses an undocumented REST API endpoint. See ScriptConfig for details.
func (c *Client) ScriptGet(ctx context.Context, id string) (*ScriptConfig, error) {
	if id == "" {
		return nil, fmt.Errorf("script id is required")
	}

	var config ScriptConfig
	path := fmt.Sprintf("/api/config/script/config/%s", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &config); err != nil {
		return nil, fmt.Errorf("script get: %w", err)
	}
	return &config, nil
}

// ScriptSave creates or updates a script configuration.
//
// WARNING: This uses an undocumented REST API endpoint. See ScriptConfig for details.
func (c *Client) ScriptSave(ctx context.Context, config *ScriptConfig) error {
	if config == nil {
		return fmt.Errorf("script config is required")
	}
	if config.ID == "" {
		return fmt.Errorf("script id is required")
	}
	if config.Alias == "" {
		return fmt.Errorf("script alias is required")
	}
	if len(config.Sequence) == 0 {
		return fmt.Errorf("script sequence is required")
	}

	path := fmt.Sprintf("/api/config/script/config/%s", config.ID)
	if err := c.doJSON(ctx, http.MethodPost, path, config, nil); err != nil {
		return fmt.Errorf("script save: %w", err)
	}
	return nil
}

// ScriptDeleteConfig deletes a script configuration by ID.
//
// WARNING: This uses an undocumented REST API endpoint. See ScriptConfig for details.
func (c *Client) ScriptDeleteConfig(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("script id is required")
	}

	path := fmt.Sprintf("/api/config/script/config/%s", id)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("script delete: %w", err)
	}
	return nil
}

// ScriptReload reloads all scripts from YAML configuration.
func (c *Client) ScriptReload(ctx context.Context) error {
	_, err := c.CallService(ctx, "script", "reload", nil)
	if err != nil {
		return fmt.Errorf("script reload: %w", err)
	}
	return nil
}

// ScriptRun executes a script by calling it as a service.
// Variables are passed as flattened service data.
//
// The script name is extracted from the entity_id (e.g., "script.test" → "test")
// and called as script.{name} service.
func (c *Client) ScriptRun(ctx context.Context, entityID string, variables map[string]any) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	// Extract script name from entity_id
	scriptName := strings.TrimPrefix(entityID, "script.")
	if scriptName == entityID {
		return fmt.Errorf("invalid entity_id format: expected 'script.*', got '%s'", entityID)
	}

	// Call the script as a service with variables as service data
	_, err := c.CallService(ctx, "script", scriptName, &ServiceCallRequest{
		Data: variables,
	})
	if err != nil {
		return fmt.Errorf("script run: %w", err)
	}
	return nil
}

// ScriptTurnOn turns on a script using the script.turn_on service.
// Variables are nested under a "variables" key in the service call.
//
// This is an alternative to ScriptRun() that supports asynchronous execution.
func (c *Client) ScriptTurnOn(ctx context.Context, entityID string, variables map[string]any) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	req := &ServiceCallRequest{
		EntityID: entityID,
	}

	// If variables are provided, nest them under "variables" key
	if len(variables) > 0 {
		req.Data = map[string]any{
			"variables": variables,
		}
	}

	_, err := c.CallService(ctx, "script", "turn_on", req)
	if err != nil {
		return fmt.Errorf("script turn_on: %w", err)
	}
	return nil
}

// ScriptTurnOff stops a running script.
func (c *Client) ScriptTurnOff(ctx context.Context, entityID string) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	_, err := c.CallService(ctx, "script", "turn_off", &ServiceCallRequest{
		EntityID: entityID,
	})
	if err != nil {
		return fmt.Errorf("script turn_off: %w", err)
	}
	return nil
}

// ScriptToggle toggles a script's running state.
func (c *Client) ScriptToggle(ctx context.Context, entityID string) error {
	if entityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	_, err := c.CallService(ctx, "script", "toggle", &ServiceCallRequest{
		EntityID: entityID,
	})
	if err != nil {
		return fmt.Errorf("script toggle: %w", err)
	}
	return nil
}
