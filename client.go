package hago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the Home Assistant REST API client.
// It is safe for concurrent use by multiple goroutines.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option is a functional option for configuring the Client.
type Option func(*Client) error

// WithBaseURL sets the base URL for the Home Assistant instance.
// The URL should include the protocol and port, e.g., "http://homeassistant.local:8123".
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		// Normalize URL - remove trailing slash
		c.baseURL = strings.TrimSuffix(baseURL, "/")
		return nil
	}
}

// WithToken sets the Long-Lived Access Token for authentication.
func WithToken(token string) Option {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

// WithTimeout sets the timeout for HTTP requests.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

// New creates a new Home Assistant API client with the given options.
// At minimum, WithBaseURL and WithToken must be provided.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// validate checks that required fields are set.
func (c *Client) validate() error {
	if c.baseURL == "" {
		return ErrNoBaseURL
	}
	if c.token == "" {
		return ErrNoToken
	}
	return nil
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// doRequest performs an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, &RequestError{Op: "marshal request body", Err: err}
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, &RequestError{Op: "create request", Err: err}
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &RequestError{Op: "execute request", Err: err}
	}

	return resp, nil
}

// doJSON performs an HTTP request and decodes the JSON response.
func (c *Client) doJSON(ctx context.Context, method, path string, body any, result any) error {
	resp, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return &RequestError{Op: "decode response", Err: err}
		}
	}

	return nil
}

// doText performs an HTTP request and returns the response as text.
func (c *Client) doText(ctx context.Context, method, path string, body any) (string, error) {
	resp, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return "", err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &RequestError{Op: "read response body", Err: err}
	}

	return string(data), nil
}

// doBytes performs an HTTP request and returns the response as bytes.
func (c *Client) doBytes(ctx context.Context, method, path string, body any) ([]byte, error) {
	resp, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &RequestError{Op: "read response body", Err: err}
	}

	return data, nil
}

// checkResponse checks the HTTP response for errors.
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(body),
	}

	// Try to parse error message from JSON response
	var errResp struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
		apiErr.Message = errResp.Message
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest:
		if apiErr.Message != "" {
			return apiErr
		}
		return ErrBadRequest
	case http.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	default:
		return apiErr
	}
}

// buildQueryString builds a query string from parameters.
func buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Set(k, v)
		}
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

// formatTime formats a time for the API.
func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// formatTimePtr formats a time pointer for the API, returning empty string if nil.
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}

// boolToString converts a bool to "true" or empty string.
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return ""
}

// Status checks if the API is running.
func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	var result StatusResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Config returns the current Home Assistant configuration.
func (c *Client) Config(ctx context.Context) (*Config, error) {
	var result Config
	if err := c.doJSON(ctx, http.MethodGet, "/api/config", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Components returns a list of loaded components.
func (c *Client) Components(ctx context.Context) ([]string, error) {
	var result []string
	if err := c.doJSON(ctx, http.MethodGet, "/api/components", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Events returns a list of events with listener counts.
func (c *Client) Events(ctx context.Context) ([]Event, error) {
	var result []Event
	if err := c.doJSON(ctx, http.MethodGet, "/api/events", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FireEvent fires an event with optional data.
func (c *Client) FireEvent(ctx context.Context, eventType string, data EventData) error {
	path := fmt.Sprintf("/api/events/%s", eventType)
	return c.doJSON(ctx, http.MethodPost, path, data, nil)
}

// Services returns a list of available services grouped by domain.
func (c *Client) Services(ctx context.Context) ([]Service, error) {
	var result []Service
	if err := c.doJSON(ctx, http.MethodGet, "/api/services", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CallService calls a service in a specific domain.
// The request can include an entity_id and additional service-specific data.
func (c *Client) CallService(ctx context.Context, domain, service string, request *ServiceCallRequest) ([]State, error) {
	path := fmt.Sprintf("/api/services/%s/%s", domain, service)
	var result []State
	if err := c.doJSON(ctx, http.MethodPost, path, request, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// States returns a list of all entity states.
func (c *Client) States(ctx context.Context) ([]State, error) {
	var result []State
	if err := c.doJSON(ctx, http.MethodGet, "/api/states", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// State returns the state of a specific entity.
func (c *Client) State(ctx context.Context, entityID string) (*State, error) {
	path := fmt.Sprintf("/api/states/%s", entityID)
	var result State
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetState creates or updates the state of an entity.
// Returns the resulting state.
func (c *Client) SetState(ctx context.Context, entityID string, update *StateUpdate) (*State, error) {
	path := fmt.Sprintf("/api/states/%s", entityID)
	var result State
	if err := c.doJSON(ctx, http.MethodPost, path, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteState deletes an entity from Home Assistant.
func (c *Client) DeleteState(ctx context.Context, entityID string) error {
	path := fmt.Sprintf("/api/states/%s", entityID)
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

// History returns state history for entities.
// The timestamp parameter specifies the start time for the history.
// Options can filter and customize the response.
func (c *Client) History(ctx context.Context, timestamp time.Time, opts *HistoryOptions) ([][]HistoryEntry, error) {
	params := make(map[string]string)
	if opts != nil {
		params["filter_entity_id"] = opts.FilterEntityID
		params["end_time"] = formatTimePtr(opts.EndTime)
		params["minimal_response"] = boolToString(opts.MinimalResponse)
		params["no_attributes"] = boolToString(opts.NoAttributes)
		params["significant_changes_only"] = boolToString(opts.SignificantChangesOnly)
	}

	path := fmt.Sprintf("/api/history/period/%s%s", formatTime(timestamp), buildQueryString(params))
	var result [][]HistoryEntry
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Logbook returns logbook entries starting from a timestamp.
func (c *Client) Logbook(ctx context.Context, timestamp time.Time, opts *LogbookOptions) ([]LogbookEntry, error) {
	params := make(map[string]string)
	if opts != nil {
		params["entity"] = opts.Entity
		params["end_time"] = formatTimePtr(opts.EndTime)
	}

	path := fmt.Sprintf("/api/logbook/%s%s", formatTime(timestamp), buildQueryString(params))
	var result []LogbookEntry
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ErrorLog returns the Home Assistant error log as plain text.
func (c *Client) ErrorLog(ctx context.Context) (string, error) {
	return c.doText(ctx, http.MethodGet, "/api/error_log", nil)
}

// RenderTemplate renders a Jinja2 template and returns the result.
func (c *Client) RenderTemplate(ctx context.Context, template string) (string, error) {
	req := &TemplateRequest{Template: template}
	return c.doText(ctx, http.MethodPost, "/api/template", req)
}

// CheckConfig validates the Home Assistant configuration.
func (c *Client) CheckConfig(ctx context.Context) (*ConfigCheckResult, error) {
	var result ConfigCheckResult
	if err := c.doJSON(ctx, http.MethodPost, "/api/config/core/check_config", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CameraProxy returns a JPEG image from a camera entity.
func (c *Client) CameraProxy(ctx context.Context, entityID string) ([]byte, error) {
	path := fmt.Sprintf("/api/camera_proxy/%s", entityID)
	return c.doBytes(ctx, http.MethodGet, path, nil)
}

// Calendars returns a list of calendar entities.
func (c *Client) Calendars(ctx context.Context) ([]Calendar, error) {
	var result []Calendar
	if err := c.doJSON(ctx, http.MethodGet, "/api/calendars", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CalendarEvents returns events for a calendar entity within a time range.
func (c *Client) CalendarEvents(ctx context.Context, entityID string, start, end time.Time) ([]CalendarEvent, error) {
	params := map[string]string{
		"start": formatTime(start),
		"end":   formatTime(end),
	}
	path := fmt.Sprintf("/api/calendars/%s%s", entityID, buildQueryString(params))
	var result []CalendarEvent
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// HandleIntent processes an intent and returns the response.
func (c *Client) HandleIntent(ctx context.Context, intent *IntentRequest) (*IntentResponse, error) {
	var result IntentResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/intent/handle", intent, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
