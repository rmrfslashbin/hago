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

func TestRequestError_Error(t *testing.T) {
	err := &RequestError{Op: "test operation", Err: ErrNoBaseURL}
	want := "test operation: base URL is required"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestClient_Components(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/components" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"homeassistant", "light", "switch", "sensor"})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	components, err := client.Components(context.Background())
	if err != nil {
		t.Fatalf("Components() error = %v", err)
	}

	if len(components) != 4 {
		t.Errorf("Components() returned %d components, want 4", len(components))
	}
}

func TestClient_Services(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/services" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Service{
			{Domain: "light", Services: map[string]ServiceDetails{"turn_on": {Name: "Turn on"}}},
			{Domain: "switch", Services: map[string]ServiceDetails{"toggle": {Name: "Toggle"}}},
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

	services, err := client.Services(context.Background())
	if err != nil {
		t.Fatalf("Services() error = %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Services() returned %d services, want 2", len(services))
	}
	if services[0].Domain != "light" {
		t.Errorf("Services()[0].Domain = %v, want %v", services[0].Domain, "light")
	}
}

func TestClient_FireEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/api/events/test_event" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var data map[string]any
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event test_event fired."})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = client.FireEvent(context.Background(), "test_event", EventData{"key": "value"})
	if err != nil {
		t.Fatalf("FireEvent() error = %v", err)
	}
}

func TestClient_SetState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/api/states/sensor.test" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var update StateUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(State{
			EntityID:   "sensor.test",
			State:      update.State,
			Attributes: update.Attributes,
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

	state, err := client.SetState(context.Background(), "sensor.test", &StateUpdate{
		State:      "42",
		Attributes: map[string]any{"unit": "celsius"},
	})
	if err != nil {
		t.Fatalf("SetState() error = %v", err)
	}

	if state.State != "42" {
		t.Errorf("SetState().State = %v, want %v", state.State, "42")
	}
}

func TestClient_DeleteState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/api/states/sensor.test" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = client.DeleteState(context.Background(), "sensor.test")
	if err != nil {
		t.Fatalf("DeleteState() error = %v", err)
	}
}

func TestClient_History(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check path starts with /api/history/period/
		if len(r.URL.Path) < 20 || r.URL.Path[:20] != "/api/history/period/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Check query params
		if r.URL.Query().Get("filter_entity_id") != "light.test" {
			t.Errorf("filter_entity_id = %v, want light.test", r.URL.Query().Get("filter_entity_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([][]HistoryEntry{
			{
				{EntityID: "light.test", State: "on"},
				{EntityID: "light.test", State: "off"},
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

	history, err := client.History(context.Background(), time.Now().Add(-24*time.Hour), &HistoryOptions{
		FilterEntityID: "light.test",
	})
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}

	if len(history) != 1 {
		t.Errorf("History() returned %d entity histories, want 1", len(history))
	}
	if len(history[0]) != 2 {
		t.Errorf("History()[0] returned %d entries, want 2", len(history[0]))
	}
}

func TestClient_Logbook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check path starts with /api/logbook/
		if len(r.URL.Path) < 13 || r.URL.Path[:13] != "/api/logbook/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]LogbookEntry{
			{Name: "Light", Message: "turned on"},
			{Name: "Switch", Message: "turned off"},
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

	entries, err := client.Logbook(context.Background(), time.Now().Add(-24*time.Hour), nil)
	if err != nil {
		t.Fatalf("Logbook() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Logbook() returned %d entries, want 2", len(entries))
	}
}

func TestClient_ErrorLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/error_log" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("2024-01-01 12:00:00 ERROR Something went wrong\n"))
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log, err := client.ErrorLog(context.Background())
	if err != nil {
		t.Fatalf("ErrorLog() error = %v", err)
	}

	if log == "" {
		t.Error("ErrorLog() returned empty string")
	}
}

func TestClient_CheckConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost || r.URL.Path != "/api/config/core/check_config" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ConfigCheckResult{
			Result: "valid",
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

	result, err := client.CheckConfig(context.Background())
	if err != nil {
		t.Fatalf("CheckConfig() error = %v", err)
	}

	if result.Result != "valid" {
		t.Errorf("CheckConfig().Result = %v, want %v", result.Result, "valid")
	}
}

func TestClient_Calendars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/calendars" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Calendar{
			{EntityID: "calendar.personal", Name: "Personal"},
			{EntityID: "calendar.work", Name: "Work"},
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

	calendars, err := client.Calendars(context.Background())
	if err != nil {
		t.Fatalf("Calendars() error = %v", err)
	}

	if len(calendars) != 2 {
		t.Errorf("Calendars() returned %d calendars, want 2", len(calendars))
	}
}

func TestClient_CalendarEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check path starts with /api/calendars/
		if len(r.URL.Path) < 15 || r.URL.Path[:15] != "/api/calendars/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Check for start and end params
		if r.URL.Query().Get("start") == "" || r.URL.Query().Get("end") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CalendarEvent{
			{Summary: "Meeting", Start: "2024-01-01T10:00:00", End: "2024-01-01T11:00:00"},
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

	start := time.Now()
	end := start.AddDate(0, 0, 7)
	events, err := client.CalendarEvents(context.Background(), "calendar.personal", start, end)
	if err != nil {
		t.Fatalf("CalendarEvents() error = %v", err)
	}

	if len(events) != 1 {
		t.Errorf("CalendarEvents() returned %d events, want 1", len(events))
	}
}

func TestClient_CameraProxy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/camera_proxy/camera.front_door" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // JPEG magic bytes
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	data, err := client.CameraProxy(context.Background(), "camera.front_door")
	if err != nil {
		t.Fatalf("CameraProxy() error = %v", err)
	}

	if len(data) != 4 {
		t.Errorf("CameraProxy() returned %d bytes, want 4", len(data))
	}
}

func TestClient_HandleIntent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost || r.URL.Path != "/api/intent/handle" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req IntentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(IntentResponse{
			Speech: SpeechResponse{
				Plain: PlainSpeech{Speech: "Turned on the light"},
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

	resp, err := client.HandleIntent(context.Background(), &IntentRequest{
		Name: "HassTurnOn",
		Data: map[string]any{"name": "living room light"},
	})
	if err != nil {
		t.Fatalf("HandleIntent() error = %v", err)
	}

	if resp.Speech.Plain.Speech != "Turned on the light" {
		t.Errorf("HandleIntent().Speech = %v, want %v", resp.Speech.Plain.Speech, "Turned on the light")
	}
}

func TestClient_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	client, err := New(
		WithBaseURL("http://localhost:8123"),
		WithToken("test-token"),
		WithHTTPClient(customClient),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.httpClient != customClient {
		t.Error("WithHTTPClient() did not set custom client")
	}
}

func TestClient_APIErrorWithMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid entity ID format"})
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
	if err == nil {
		t.Fatal("Status() should return error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Error should be *APIError, got %T", err)
	}

	if apiErr.Message != "Invalid entity ID format" {
		t.Errorf("APIError.Message = %v, want %v", apiErr.Message, "Invalid entity ID format")
	}
}

func TestClient_GenericAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
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
	if err == nil {
		t.Fatal("Status() should return error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Error should be *APIError, got %T", err)
	}

	if apiErr.StatusCode != 500 {
		t.Errorf("APIError.StatusCode = %v, want %v", apiErr.StatusCode, 500)
	}
}

func TestBuildQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name:   "empty params",
			params: map[string]string{},
			want:   "",
		},
		{
			name:   "nil-like empty values",
			params: map[string]string{"key": ""},
			want:   "",
		},
		{
			name:   "single param",
			params: map[string]string{"key": "value"},
			want:   "?key=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildQueryString(tt.params)
			if tt.want == "" && got != "" {
				t.Errorf("buildQueryString() = %v, want empty", got)
			}
			if tt.want != "" && got == "" {
				t.Errorf("buildQueryString() = empty, want %v", tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got := formatTime(tm)
	want := "2024-01-15T10:30:00Z"
	if got != want {
		t.Errorf("formatTime() = %v, want %v", got, want)
	}
}

func TestFormatTimePtr(t *testing.T) {
	// Nil case
	if got := formatTimePtr(nil); got != "" {
		t.Errorf("formatTimePtr(nil) = %v, want empty", got)
	}

	// Non-nil case
	tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got := formatTimePtr(&tm)
	want := "2024-01-15T10:30:00Z"
	if got != want {
		t.Errorf("formatTimePtr() = %v, want %v", got, want)
	}
}

func TestBoolToString(t *testing.T) {
	if got := boolToString(true); got != "true" {
		t.Errorf("boolToString(true) = %v, want true", got)
	}
	if got := boolToString(false); got != "" {
		t.Errorf("boolToString(false) = %v, want empty", got)
	}
}

func TestHistoryOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check query params
		q := r.URL.Query()
		if q.Get("minimal_response") != "true" {
			t.Errorf("minimal_response = %v, want true", q.Get("minimal_response"))
		}
		if q.Get("no_attributes") != "true" {
			t.Errorf("no_attributes = %v, want true", q.Get("no_attributes"))
		}
		if q.Get("significant_changes_only") != "true" {
			t.Errorf("significant_changes_only = %v, want true", q.Get("significant_changes_only"))
		}
		if q.Get("filter_entity_id") != "sensor.test" {
			t.Errorf("filter_entity_id = %v, want sensor.test", q.Get("filter_entity_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([][]HistoryEntry{})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	endTime := time.Now()
	_, err = client.History(context.Background(), time.Now().Add(-24*time.Hour), &HistoryOptions{
		FilterEntityID:         "sensor.test",
		EndTime:                &endTime,
		MinimalResponse:        true,
		NoAttributes:           true,
		SignificantChangesOnly: true,
	})
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
}

func TestLogbookOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check query params
		q := r.URL.Query()
		if q.Get("entity") != "light.test" {
			t.Errorf("entity = %v, want light.test", q.Get("entity"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]LogbookEntry{})
	}))
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	endTime := time.Now()
	_, err = client.Logbook(context.Background(), time.Now().Add(-24*time.Hour), &LogbookOptions{
		Entity:  "light.test",
		EndTime: &endTime,
	})
	if err != nil {
		t.Fatalf("Logbook() error = %v", err)
	}
}
