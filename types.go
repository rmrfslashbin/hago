package hago

import (
	"encoding/json"
	"time"
)

// StatusResponse represents the response from the API status endpoint.
type StatusResponse struct {
	Message string `json:"message"`
}

// Config represents the Home Assistant configuration.
type Config struct {
	Components      []string `json:"components"`
	ConfigDir       string   `json:"config_dir"`
	Elevation       int      `json:"elevation"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	LocationName    string   `json:"location_name"`
	TimeZone        string   `json:"time_zone"`
	UnitSystem      UnitSystem `json:"unit_system"`
	Version         string   `json:"version"`
	State           string   `json:"state"`
	ExternalURL     string   `json:"external_url,omitempty"`
	InternalURL     string   `json:"internal_url,omitempty"`
	Currency        string   `json:"currency,omitempty"`
	SafeMode        bool     `json:"safe_mode"`
	AllowlistExternalDirs []string `json:"allowlist_external_dirs,omitempty"`
	AllowlistExternalURLs []string `json:"allowlist_external_urls,omitempty"`
}

// UnitSystem represents the unit system configuration.
type UnitSystem struct {
	Length      string `json:"length"`
	Mass        string `json:"mass"`
	Temperature string `json:"temperature"`
	Volume      string `json:"volume"`
	Pressure    string `json:"pressure"`
	WindSpeed   string `json:"wind_speed"`
	Accumulated string `json:"accumulated_precipitation"`
}

// State represents the state of an entity.
type State struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]any         `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
	Context     Context                `json:"context"`
}

// Context represents the context of a state change.
type Context struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// StateUpdate represents a request to update an entity's state.
type StateUpdate struct {
	State      string         `json:"state"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Event represents an event type in Home Assistant.
type Event struct {
	Event         string `json:"event"`
	ListenerCount int    `json:"listener_count"`
}

// EventData represents data to fire with an event.
type EventData map[string]any

// Service represents a service available in Home Assistant.
type Service struct {
	Domain   string                    `json:"domain"`
	Services map[string]ServiceDetails `json:"services"`
}

// ServiceDetails contains the details of a specific service.
type ServiceDetails struct {
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Fields      map[string]ServiceField  `json:"fields,omitempty"`
	Target      *ServiceTarget           `json:"target,omitempty"`
}

// ServiceField represents a field in a service call.
type ServiceField struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Example     any         `json:"example,omitempty"`
	Default     any         `json:"default,omitempty"`
	Selector    any         `json:"selector,omitempty"`
}

// ServiceTarget represents the target specification for a service.
type ServiceTarget struct {
	Entity []TargetSelector `json:"entity,omitempty"`
	Device []TargetSelector `json:"device,omitempty"`
	Area   []TargetSelector `json:"area,omitempty"`
}

// TargetSelector represents a target selector specification.
type TargetSelector struct {
	Integration string   `json:"integration,omitempty"`
	Domain      []string `json:"domain,omitempty"`
}

// ServiceCallRequest represents a request to call a service.
type ServiceCallRequest struct {
	EntityID string         `json:"entity_id,omitempty"`
	Data     map[string]any `json:"-"`
}

// MarshalJSON implements custom JSON marshaling to flatten the Data field.
func (s *ServiceCallRequest) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	if s.EntityID != "" {
		m["entity_id"] = s.EntityID
	}
	for k, v := range s.Data {
		m[k] = v
	}
	return json.Marshal(m)
}

// HistoryEntry represents a historical state entry.
type HistoryEntry struct {
	EntityID    string    `json:"entity_id"`
	State       string    `json:"state"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	LastChanged time.Time `json:"last_changed"`
	LastUpdated time.Time `json:"last_updated"`
}

// HistoryOptions contains options for history queries.
type HistoryOptions struct {
	// FilterEntityID filters history to specific entities (comma-separated).
	FilterEntityID string
	// EndTime limits results to before this time.
	EndTime *time.Time
	// MinimalResponse returns only last_changed and state.
	MinimalResponse bool
	// NoAttributes excludes attributes from the response.
	NoAttributes bool
	// SignificantChangesOnly returns only significant state changes.
	SignificantChangesOnly bool
}

// LogbookEntry represents an entry in the logbook.
type LogbookEntry struct {
	When      time.Time `json:"when"`
	Name      string    `json:"name"`
	Message   string    `json:"message,omitempty"`
	Domain    string    `json:"domain,omitempty"`
	EntityID  string    `json:"entity_id,omitempty"`
	State     string    `json:"state,omitempty"`
	Icon      string    `json:"icon,omitempty"`
	ContextID string    `json:"context_id,omitempty"`
}

// LogbookOptions contains options for logbook queries.
type LogbookOptions struct {
	// Entity filters to a specific entity.
	Entity string
	// EndTime limits results to before this time.
	EndTime *time.Time
}

// TemplateRequest represents a template rendering request.
type TemplateRequest struct {
	Template string `json:"template"`
}

// ConfigCheckResult represents the result of a configuration check.
type ConfigCheckResult struct {
	Result string `json:"result"`
	Errors string `json:"errors,omitempty"`
}

// Calendar represents a calendar entity.
type Calendar struct {
	EntityID string `json:"entity_id"`
	Name     string `json:"name"`
}

// CalendarEvent represents an event in a calendar.
type CalendarEvent struct {
	Start       string         `json:"start"`
	End         string         `json:"end"`
	Summary     string         `json:"summary"`
	Description string         `json:"description,omitempty"`
	Location    string         `json:"location,omitempty"`
	UID         string         `json:"uid,omitempty"`
	Recurrence  string         `json:"recurrence_id,omitempty"`
	RRULE       string         `json:"rrule,omitempty"`
}

// IntentRequest represents a request to handle an intent.
type IntentRequest struct {
	Name string         `json:"name"`
	Data map[string]any `json:"data,omitempty"`
}

// IntentResponse represents the response from an intent.
type IntentResponse struct {
	Speech     SpeechResponse `json:"speech,omitempty"`
	Card       CardResponse   `json:"card,omitempty"`
	LanguageCode string       `json:"language,omitempty"`
}

// SpeechResponse represents the speech portion of an intent response.
type SpeechResponse struct {
	Plain PlainSpeech `json:"plain,omitempty"`
}

// PlainSpeech represents plain speech output.
type PlainSpeech struct {
	Speech string `json:"speech"`
}

// CardResponse represents a card in an intent response.
type CardResponse struct {
	Simple SimpleCard `json:"simple,omitempty"`
}

// SimpleCard represents a simple card response.
type SimpleCard struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
