package hago

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr error
	}{
		{
			name:    "missing base URL",
			opts:    []Option{WithToken("test-token")},
			wantErr: ErrNoBaseURL,
		},
		{
			name:    "missing token",
			opts:    []Option{WithBaseURL("http://localhost:8123")},
			wantErr: ErrNoToken,
		},
		{
			name: "valid config",
			opts: []Option{
				WithBaseURL("http://localhost:8123"),
				WithToken("test-token"),
			},
			wantErr: nil,
		},
		{
			name: "with trailing slash",
			opts: []Option{
				WithBaseURL("http://localhost:8123/"),
				WithToken("test-token"),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.opts...)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("New() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}
			if client == nil {
				t.Error("New() returned nil client")
			}
		})
	}
}

func TestClient_BaseURL(t *testing.T) {
	client, err := New(
		WithBaseURL("http://localhost:8123/"),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should strip trailing slash
	if got := client.BaseURL(); got != "http://localhost:8123" {
		t.Errorf("BaseURL() = %v, want %v", got, "http://localhost:8123")
	}
}

func TestClient_Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StatusResponse{Message: "API running."})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if status.Message != "API running." {
		t.Errorf("Status().Message = %v, want %v", status.Message, "API running.")
	}
}

func TestClient_Config(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/config" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Config{
			LocationName: "Home",
			Version:      "2024.1.0",
			TimeZone:     "America/New_York",
		})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	config, err := client.Config(context.Background())
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}

	if config.LocationName != "Home" {
		t.Errorf("Config().LocationName = %v, want %v", config.LocationName, "Home")
	}
	if config.Version != "2024.1.0" {
		t.Errorf("Config().Version = %v, want %v", config.Version, "2024.1.0")
	}
}

func TestClient_States(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/states" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]State{
			{EntityID: "light.living_room", State: "on"},
			{EntityID: "switch.bedroom", State: "off"},
		})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	states, err := client.States(context.Background())
	if err != nil {
		t.Fatalf("States() error = %v", err)
	}

	if len(states) != 2 {
		t.Errorf("States() returned %d states, want 2", len(states))
	}
	if states[0].EntityID != "light.living_room" {
		t.Errorf("States()[0].EntityID = %v, want %v", states[0].EntityID, "light.living_room")
	}
}

func TestClient_State(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/states/light.living_room" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(State{
			EntityID: "light.living_room",
			State:    "on",
			Attributes: map[string]any{
				"brightness": 255,
				"friendly_name": "Living Room Light",
			},
		})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	state, err := client.State(context.Background(), "light.living_room")
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}

	if state.EntityID != "light.living_room" {
		t.Errorf("State().EntityID = %v, want %v", state.EntityID, "light.living_room")
	}
	if state.State != "on" {
		t.Errorf("State().State = %v, want %v", state.State, "on")
	}
}

func TestClient_CallService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/api/services/light/turn_on" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Verify request body
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req["entity_id"] != "light.living_room" {
			t.Errorf("entity_id = %v, want %v", req["entity_id"], "light.living_room")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]State{
			{EntityID: "light.living_room", State: "on"},
		})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	states, err := client.CallService(context.Background(), "light", "turn_on", &ServiceCallRequest{
		EntityID: "light.living_room",
	})
	if err != nil {
		t.Fatalf("CallService() error = %v", err)
	}

	if len(states) != 1 {
		t.Errorf("CallService() returned %d states, want 1", len(states))
	}
}

func TestClient_Events(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/events" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Event{
			{Event: "state_changed", ListenerCount: 5},
			{Event: "call_service", ListenerCount: 2},
		})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	events, err := client.Events(context.Background())
	if err != nil {
		t.Fatalf("Events() error = %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Events() returned %d events, want 2", len(events))
	}
}

func TestClient_RenderTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost || r.URL.Path != "/api/template" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req TemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return plain text (not JSON)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := client.RenderTemplate(context.Background(), "{{ 'Hello, World!' }}")
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}

	if result != "Hello, World!" {
		t.Errorf("RenderTemplate() = %v, want %v", result, "Hello, World!")
	}
}

func TestClient_ErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    ErrUnauthorized,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			wantErr:    ErrNotFound,
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			wantErr:    ErrBadRequest,
		},
		{
			name:       "method not allowed",
			statusCode: http.StatusMethodNotAllowed,
			wantErr:    ErrMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, err := New(
				WithBaseURL(server.URL),
				WithToken("test-token"),
			)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			_, err = client.Status(context.Background())
			if err != tt.wantErr {
				t.Errorf("Status() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_WithTimeout(t *testing.T) {
	client, err := New(
		WithBaseURL("http://localhost:8123"),
		WithToken("test-token"),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, 5*time.Second)
	}
}

func TestServiceCallRequest_MarshalJSON(t *testing.T) {
	req := &ServiceCallRequest{
		EntityID: "light.living_room",
		Data: map[string]any{
			"brightness": 255,
			"transition": 2,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result["entity_id"] != "light.living_room" {
		t.Errorf("entity_id = %v, want %v", result["entity_id"], "light.living_room")
	}
	if result["brightness"] != float64(255) {
		t.Errorf("brightness = %v, want %v", result["brightness"], 255)
	}
	if result["transition"] != float64(2) {
		t.Errorf("transition = %v, want %v", result["transition"], 2)
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *APIError
		wantStr string
	}{
		{
			name: "with message",
			err: &APIError{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "Invalid entity ID",
			},
			wantStr: "API error 400 (Bad Request): Invalid entity ID",
		},
		{
			name: "with body",
			err: &APIError{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Body:       "Something went wrong",
			},
			wantStr: "API error 500 (Internal Server Error): Something went wrong",
		},
		{
			name: "minimal",
			err: &APIError{
				StatusCode: 503,
				Status:     "Service Unavailable",
			},
			wantStr: "API error 503 (Service Unavailable)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantStr {
				t.Errorf("Error() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestRequestError_Unwrap(t *testing.T) {
	inner := ErrNoBaseURL
	err := &RequestError{Op: "test", Err: inner}

	if err.Unwrap() != inner {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), inner)
	}
}
