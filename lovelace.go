package hago

import (
	"context"
	"encoding/json"
)

// Dashboard represents a Lovelace dashboard.
type Dashboard struct {
	ID              string `json:"id,omitempty"`
	URLPath         string `json:"url_path"`
	Title           string `json:"title,omitempty"`
	Icon            string `json:"icon,omitempty"`
	ShowInSidebar   bool   `json:"show_in_sidebar,omitempty"`
	RequireAdmin    bool   `json:"require_admin,omitempty"`
	Mode            string `json:"mode,omitempty"`
	AllowSingleWord bool   `json:"allow_single_word,omitempty"`
}

// DashboardConfig represents the configuration of a Lovelace dashboard.
// This is the actual dashboard content with views, cards, etc.
type DashboardConfig struct {
	Title      string         `json:"title,omitempty"`
	Views      []View         `json:"views,omitempty"`
	Strategy   *Strategy      `json:"strategy,omitempty"`
	Background string         `json:"background,omitempty"`
	// Raw holds the full config as received, for pass-through scenarios
	Raw json.RawMessage `json:"-"`
}

// View represents a view (tab) in a Lovelace dashboard.
type View struct {
	Title      string         `json:"title,omitempty"`
	Path       string         `json:"path,omitempty"`
	Icon       string         `json:"icon,omitempty"`
	Theme      string         `json:"theme,omitempty"`
	Panel      bool           `json:"panel,omitempty"`
	Background string         `json:"background,omitempty"`
	Badges     []any          `json:"badges,omitempty"`
	Cards      []any          `json:"cards,omitempty"`
	Subview    bool           `json:"subview,omitempty"`
	Strategy   *Strategy      `json:"strategy,omitempty"`
}

// Strategy represents a dashboard or view generation strategy.
type Strategy struct {
	Type    string         `json:"type"`
	Options map[string]any `json:"options,omitempty"`
}

// Resource represents a Lovelace resource (custom card, theme, etc).
type Resource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// CreateDashboardRequest is the request to create a new dashboard.
type CreateDashboardRequest struct {
	URLPath         string `json:"url_path"`
	Title           string `json:"title,omitempty"`
	Icon            string `json:"icon,omitempty"`
	ShowInSidebar   bool   `json:"show_in_sidebar,omitempty"`
	RequireAdmin    bool   `json:"require_admin,omitempty"`
	AllowSingleWord bool   `json:"allow_single_word,omitempty"`
}

// UpdateDashboardRequest is the request to update a dashboard.
type UpdateDashboardRequest struct {
	Title           *string `json:"title,omitempty"`
	Icon            *string `json:"icon,omitempty"`
	ShowInSidebar   *bool   `json:"show_in_sidebar,omitempty"`
	RequireAdmin    *bool   `json:"require_admin,omitempty"`
	AllowSingleWord *bool   `json:"allow_single_word,omitempty"`
}

// lovelaceConfigCmd is the WebSocket command to get a dashboard config.
type lovelaceConfigCmd struct {
	Type    string  `json:"type"`
	URLPath *string `json:"url_path,omitempty"`
	Force   bool    `json:"force,omitempty"`
}

// lovelaceSaveConfigCmd is the WebSocket command to save a dashboard config.
type lovelaceSaveConfigCmd struct {
	Type    string `json:"type"`
	URLPath *string `json:"url_path,omitempty"`
	Config  any    `json:"config"`
}

// lovelaceDeleteConfigCmd is the WebSocket command to delete a dashboard config.
type lovelaceDeleteConfigCmd struct {
	Type    string  `json:"type"`
	URLPath *string `json:"url_path,omitempty"`
}

// lovelaceDashboardsListCmd is the WebSocket command to list dashboards.
type lovelaceDashboardsListCmd struct {
	Type string `json:"type"`
}

// lovelaceDashboardsCreateCmd is the WebSocket command to create a dashboard.
type lovelaceDashboardsCreateCmd struct {
	Type            string `json:"type"`
	URLPath         string `json:"url_path"`
	Title           string `json:"title,omitempty"`
	Icon            string `json:"icon,omitempty"`
	ShowInSidebar   bool   `json:"show_in_sidebar,omitempty"`
	RequireAdmin    bool   `json:"require_admin,omitempty"`
	AllowSingleWord bool   `json:"allow_single_word,omitempty"`
}

// lovelaceDashboardsUpdateCmd is the WebSocket command to update a dashboard.
type lovelaceDashboardsUpdateCmd struct {
	Type        string `json:"type"`
	DashboardID string `json:"dashboard_id"`
	Title           *string `json:"title,omitempty"`
	Icon            *string `json:"icon,omitempty"`
	ShowInSidebar   *bool   `json:"show_in_sidebar,omitempty"`
	RequireAdmin    *bool   `json:"require_admin,omitempty"`
	AllowSingleWord *bool   `json:"allow_single_word,omitempty"`
}

// lovelaceDashboardsDeleteCmd is the WebSocket command to delete a dashboard.
type lovelaceDashboardsDeleteCmd struct {
	Type        string `json:"type"`
	DashboardID string `json:"dashboard_id"`
}

// lovelaceResourcesCmd is the WebSocket command to list resources.
type lovelaceResourcesCmd struct {
	Type string `json:"type"`
}

// LovelaceListDashboards returns a list of all Lovelace dashboards.
// This includes both storage-mode and YAML-mode dashboards.
func (c *Client) LovelaceListDashboards(ctx context.Context) ([]Dashboard, error) {
	cmd := lovelaceDashboardsListCmd{
		Type: "lovelace/dashboards/list",
	}

	var result []Dashboard
	if err := c.wsCommand(ctx, cmd, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LovelaceGetConfig returns the configuration for a dashboard.
// If urlPath is nil, returns the default (overview) dashboard config.
// Set force to true to bypass the cache.
func (c *Client) LovelaceGetConfig(ctx context.Context, urlPath *string, force bool) (json.RawMessage, error) {
	cmd := lovelaceConfigCmd{
		Type:    "lovelace/config",
		URLPath: urlPath,
		Force:   force,
	}

	var result json.RawMessage
	if err := c.wsCommand(ctx, cmd, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LovelaceGetConfigParsed returns the configuration for a dashboard as a parsed struct.
// If urlPath is nil, returns the default (overview) dashboard config.
func (c *Client) LovelaceGetConfigParsed(ctx context.Context, urlPath *string) (*DashboardConfig, error) {
	raw, err := c.LovelaceGetConfig(ctx, urlPath, false)
	if err != nil {
		return nil, err
	}

	var config DashboardConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	config.Raw = raw
	return &config, nil
}

// LovelaceSaveConfig saves the configuration for a dashboard.
// If urlPath is nil, saves to the default (overview) dashboard.
// The config can be a DashboardConfig struct, a map, or raw JSON.
// Requires admin authentication.
func (c *Client) LovelaceSaveConfig(ctx context.Context, urlPath *string, config any) error {
	cmd := lovelaceSaveConfigCmd{
		Type:    "lovelace/config/save",
		URLPath: urlPath,
		Config:  config,
	}

	return c.wsCommand(ctx, cmd, nil)
}

// LovelaceDeleteConfig deletes the configuration for a dashboard.
// This resets the dashboard to auto-generated mode.
// If urlPath is nil, deletes the default (overview) dashboard config.
// Requires admin authentication.
func (c *Client) LovelaceDeleteConfig(ctx context.Context, urlPath *string) error {
	cmd := lovelaceDeleteConfigCmd{
		Type:    "lovelace/config/delete",
		URLPath: urlPath,
	}

	return c.wsCommand(ctx, cmd, nil)
}

// LovelaceCreateDashboard creates a new Lovelace dashboard.
// Requires admin authentication.
func (c *Client) LovelaceCreateDashboard(ctx context.Context, req *CreateDashboardRequest) (*Dashboard, error) {
	cmd := lovelaceDashboardsCreateCmd{
		Type:            "lovelace/dashboards/create",
		URLPath:         req.URLPath,
		Title:           req.Title,
		Icon:            req.Icon,
		ShowInSidebar:   req.ShowInSidebar,
		RequireAdmin:    req.RequireAdmin,
		AllowSingleWord: req.AllowSingleWord,
	}

	var result Dashboard
	if err := c.wsCommand(ctx, cmd, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LovelaceUpdateDashboard updates an existing dashboard's metadata.
// Requires admin authentication.
func (c *Client) LovelaceUpdateDashboard(ctx context.Context, dashboardID string, req *UpdateDashboardRequest) (*Dashboard, error) {
	cmd := lovelaceDashboardsUpdateCmd{
		Type:            "lovelace/dashboards/update",
		DashboardID:     dashboardID,
		Title:           req.Title,
		Icon:            req.Icon,
		ShowInSidebar:   req.ShowInSidebar,
		RequireAdmin:    req.RequireAdmin,
		AllowSingleWord: req.AllowSingleWord,
	}

	var result Dashboard
	if err := c.wsCommand(ctx, cmd, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LovelaceDeleteDashboard deletes a dashboard entirely.
// Requires admin authentication.
func (c *Client) LovelaceDeleteDashboard(ctx context.Context, dashboardID string) error {
	cmd := lovelaceDashboardsDeleteCmd{
		Type:        "lovelace/dashboards/delete",
		DashboardID: dashboardID,
	}

	return c.wsCommand(ctx, cmd, nil)
}

// LovelaceListResources returns a list of registered Lovelace resources.
// Resources are custom cards, themes, and other frontend extensions.
func (c *Client) LovelaceListResources(ctx context.Context) ([]Resource, error) {
	cmd := lovelaceResourcesCmd{
		Type: "lovelace/resources",
	}

	var result []Resource
	if err := c.wsCommand(ctx, cmd, &result); err != nil {
		return nil, err
	}
	return result, nil
}
