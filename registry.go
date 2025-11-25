package hago

import (
	"context"
	"fmt"
)

// EntityRegistryEntry represents an entity in the registry.
type EntityRegistryEntry struct {
	EntityID      string            `json:"entity_id"`
	Name          *string           `json:"name"`
	AreaID        *string           `json:"area_id"`
	DeviceID      *string           `json:"device_id"`
	Labels        []string          `json:"labels"`
	Icon          *string           `json:"icon"`
	DisabledBy    *string           `json:"disabled_by"`
	HiddenBy      *string           `json:"hidden_by"`
	HasEntityName bool              `json:"has_entity_name"`
	Platform      string            `json:"platform"`
	Categories    map[string]string `json:"categories,omitempty"`
	OriginalIcon  *string           `json:"original_icon,omitempty"`
	OriginalName  *string           `json:"original_name,omitempty"`
	UniqueID      string            `json:"unique_id"`
}

// DeviceRegistryEntry represents a device in the registry.
type DeviceRegistryEntry struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	AreaID           *string    `json:"area_id"`
	ConfigEntries    []string   `json:"config_entries"`
	Connections      [][]string `json:"connections"`
	DisabledBy       *string    `json:"disabled_by"`
	Identifiers      [][]string `json:"identifiers"`
	Manufacturer     *string    `json:"manufacturer"`
	Model            *string    `json:"model"`
	NameByUser       *string    `json:"name_by_user"`
	SWVersion        *string    `json:"sw_version"`
	HWVersion        *string    `json:"hw_version"`
	SerialNumber     *string    `json:"serial_number"`
	ViaDeviceID      *string    `json:"via_device_id"`
	ConfigurationURL *string    `json:"configuration_url"`
	EntryType        *string    `json:"entry_type"`
	Labels           []string   `json:"labels"`
}

// AreaRegistryEntry represents an area in the registry.
type AreaRegistryEntry struct {
	AreaID  string   `json:"area_id"`
	Name    string   `json:"name"`
	FloorID *string  `json:"floor_id"`
	Icon    *string  `json:"icon"`
	Picture *string  `json:"picture"`
	Aliases []string `json:"aliases"`
	Labels  []string `json:"labels"`
}

// LabelRegistryEntry represents a label in the registry.
type LabelRegistryEntry struct {
	LabelID     string  `json:"label_id"`
	Name        string  `json:"name"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
}

// FloorRegistryEntry represents a floor in the registry.
type FloorRegistryEntry struct {
	FloorID string   `json:"floor_id"`
	Name    string   `json:"name"`
	Icon    *string  `json:"icon"`
	Level   *int     `json:"level"`
	Aliases []string `json:"aliases"`
}

// EntityRegistry lists all entities in the registry with metadata.
func (c *Client) EntityRegistry(ctx context.Context) ([]EntityRegistryEntry, error) {
	var entries []EntityRegistryEntry
	if err := c.doJSON(ctx, "GET", "/api/config/entity_registry/list", nil, &entries); err != nil {
		return nil, fmt.Errorf("entity registry: %w", err)
	}
	return entries, nil
}

// DeviceRegistry lists all devices in the registry.
func (c *Client) DeviceRegistry(ctx context.Context) ([]DeviceRegistryEntry, error) {
	var entries []DeviceRegistryEntry
	if err := c.doJSON(ctx, "GET", "/api/config/device_registry/list", nil, &entries); err != nil {
		return nil, fmt.Errorf("device registry: %w", err)
	}
	return entries, nil
}

// AreaRegistry lists all areas in the registry.
func (c *Client) AreaRegistry(ctx context.Context) ([]AreaRegistryEntry, error) {
	var entries []AreaRegistryEntry
	if err := c.doJSON(ctx, "GET", "/api/config/area_registry/list", nil, &entries); err != nil {
		return nil, fmt.Errorf("area registry: %w", err)
	}
	return entries, nil
}

// LabelRegistry lists all labels in the registry.
func (c *Client) LabelRegistry(ctx context.Context) ([]LabelRegistryEntry, error) {
	var entries []LabelRegistryEntry
	if err := c.doJSON(ctx, "GET", "/api/config/label_registry/list", nil, &entries); err != nil {
		return nil, fmt.Errorf("label registry: %w", err)
	}
	return entries, nil
}

// FloorRegistry lists all floors in the registry.
func (c *Client) FloorRegistry(ctx context.Context) ([]FloorRegistryEntry, error) {
	var entries []FloorRegistryEntry
	if err := c.doJSON(ctx, "GET", "/api/config/floor_registry/list", nil, &entries); err != nil {
		return nil, fmt.Errorf("floor registry: %w", err)
	}
	return entries, nil
}
