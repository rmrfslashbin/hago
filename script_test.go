package hago

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ScriptRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/script/test_script" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		// Variables should be flattened in the request
		if req["key1"] != "value1" {
			t.Errorf("expected key1=value1, got %v", req["key1"])
		}
		if req["key2"] != "value2" {
			t.Errorf("expected key2=value2, got %v", req["key2"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	variables := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	err := client.ScriptRun(ctx, "script.test_script", variables)
	if err != nil {
		t.Fatalf("ScriptRun() error = %v", err)
	}
}

func TestClient_ScriptRun_NoVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/script/test_script" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.ScriptRun(ctx, "script.test_script", nil)
	if err != nil {
		t.Fatalf("ScriptRun() error = %v", err)
	}
}

func TestClient_ScriptRun_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptRun(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_ScriptRun_InvalidEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptRun(ctx, "automation.test", nil)
	if err == nil {
		t.Fatal("expected error for invalid entity_id format")
	}
	if !strings.Contains(err.Error(), "invalid entity_id format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClient_ScriptTurnOn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/script/turn_on" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "script.test" {
			t.Errorf("expected entity_id script.test, got %v", req["entity_id"])
		}

		// Variables should be nested under "variables" key
		vars, ok := req["variables"].(map[string]any)
		if !ok {
			t.Error("expected variables to be nested under 'variables' key")
		} else {
			if vars["key1"] != "value1" {
				t.Errorf("expected key1=value1, got %v", vars["key1"])
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	variables := map[string]any{
		"key1": "value1",
	}

	err := client.ScriptTurnOn(ctx, "script.test", variables)
	if err != nil {
		t.Fatalf("ScriptTurnOn() error = %v", err)
	}
}

func TestClient_ScriptTurnOn_NoVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "script.test" {
			t.Errorf("expected entity_id script.test, got %v", req["entity_id"])
		}

		// Should not have variables key when no variables provided
		if _, ok := req["variables"]; ok {
			t.Error("expected no variables key when no variables provided")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.ScriptTurnOn(ctx, "script.test", nil)
	if err != nil {
		t.Fatalf("ScriptTurnOn() error = %v", err)
	}
}

func TestClient_ScriptTurnOn_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptTurnOn(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_ScriptTurnOff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/script/turn_off" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "script.test" {
			t.Errorf("expected entity_id script.test, got %v", req["entity_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.ScriptTurnOff(ctx, "script.test")
	if err != nil {
		t.Fatalf("ScriptTurnOff() error = %v", err)
	}
}

func TestClient_ScriptTurnOff_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptTurnOff(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_ScriptToggle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/script/toggle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "script.test" {
			t.Errorf("expected entity_id script.test, got %v", req["entity_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.ScriptToggle(ctx, "script.test")
	if err != nil {
		t.Fatalf("ScriptToggle() error = %v", err)
	}
}

func TestClient_ScriptToggle_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptToggle(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_ScriptReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/script/reload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.ScriptReload(ctx)
	if err != nil {
		t.Fatalf("ScriptReload() error = %v", err)
	}
}

func TestClient_ScriptList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/script/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": "test1",
				"alias": "Test Script 1",
				"sequence": [{"service": "light.turn_on"}]
			},
			{
				"id": "test2",
				"alias": "Test Script 2",
				"description": "Test description",
				"mode": "single",
				"sequence": [{"service": "light.turn_off"}],
				"variables": {"test_var": "test_value"},
				"fields": {"field1": {"description": "Field 1"}}
			}
		]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	configs, err := client.ScriptList(ctx)
	if err != nil {
		t.Fatalf("ScriptList() error = %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	if configs[0].ID != "test1" {
		t.Errorf("expected id test1, got %s", configs[0].ID)
	}
	if configs[0].Alias != "Test Script 1" {
		t.Errorf("expected alias 'Test Script 1', got %s", configs[0].Alias)
	}
	if configs[1].Description == nil || *configs[1].Description != "Test description" {
		t.Error("expected description 'Test description'")
	}
}

func TestClient_ScriptList_FallbackToStates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/config/script/config" {
			// Simulate 404 from config endpoint
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Resource not found"}`))
			return
		}

		if r.URL.Path == "/api/states" {
			// Return states with script entities
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{
					"entity_id": "light.living_room",
					"state": "on",
					"attributes": {
						"friendly_name": "Living Room Light"
					},
					"last_changed": "2024-01-01T00:00:00Z",
					"last_updated": "2024-01-01T00:00:00Z"
				},
				{
					"entity_id": "script.test_script_1",
					"state": "off",
					"attributes": {
						"friendly_name": "Test Script 1",
						"description": "My test script",
						"mode": "single"
					},
					"last_changed": "2024-01-01T00:00:00Z",
					"last_updated": "2024-01-01T00:00:00Z"
				},
				{
					"entity_id": "script.test_script_2",
					"state": "off",
					"attributes": {
						"friendly_name": "Test Script 2"
					},
					"last_changed": "2024-01-01T00:00:00Z",
					"last_updated": "2024-01-01T00:00:00Z"
				},
				{
					"entity_id": "switch.bedroom",
					"state": "off",
					"attributes": {
						"friendly_name": "Bedroom Switch"
					},
					"last_changed": "2024-01-01T00:00:00Z",
					"last_updated": "2024-01-01T00:00:00Z"
				}
			]`))
			return
		}

		t.Errorf("unexpected path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	configs, err := client.ScriptList(ctx)
	if err != nil {
		t.Fatalf("ScriptList() error = %v", err)
	}

	// Should only return script.* entities (2), not light or switch
	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	// Verify first script
	if configs[0].ID != "script.test_script_1" {
		t.Errorf("expected id script.test_script_1, got %s", configs[0].ID)
	}
	if configs[0].Alias != "Test Script 1" {
		t.Errorf("expected alias 'Test Script 1', got %s", configs[0].Alias)
	}
	if configs[0].Description == nil || *configs[0].Description != "My test script" {
		t.Error("expected description 'My test script'")
	}
	if configs[0].Mode != "single" {
		t.Errorf("expected mode 'single', got %s", configs[0].Mode)
	}

	// Verify second script
	if configs[1].ID != "script.test_script_2" {
		t.Errorf("expected id script.test_script_2, got %s", configs[1].ID)
	}
	if configs[1].Alias != "Test Script 2" {
		t.Errorf("expected alias 'Test Script 2', got %s", configs[1].Alias)
	}

	// In fallback mode, sequence array should be empty (not nil)
	if configs[0].Sequence == nil || len(configs[0].Sequence) != 0 {
		t.Error("expected empty sequence array in fallback mode")
	}
}

func TestClient_ScriptList_FallbackNoFriendlyName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/config/script/config" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.URL.Path == "/api/states" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{
					"entity_id": "script.my_script",
					"state": "off",
					"attributes": {},
					"last_changed": "2024-01-01T00:00:00Z",
					"last_updated": "2024-01-01T00:00:00Z"
				}
			]`))
			return
		}
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	configs, err := client.ScriptList(ctx)
	if err != nil {
		t.Fatalf("ScriptList() error = %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}

	// Should use entity_id without "script." prefix as alias
	if configs[0].Alias != "my_script" {
		t.Errorf("expected alias 'my_script', got %s", configs[0].Alias)
	}
}

func TestClient_ScriptList_BothEndpointsFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Both endpoints return errors
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	_, err := client.ScriptList(ctx)
	if err == nil {
		t.Fatal("expected error when both endpoints fail")
	}

	// Error should mention both failures
	errMsg := err.Error()
	if !strings.Contains(errMsg, "config endpoint failed") {
		t.Errorf("error should mention config endpoint failure: %v", err)
	}
	if !strings.Contains(errMsg, "states fallback also failed") {
		t.Errorf("error should mention states fallback failure: %v", err)
	}
}

func TestClient_ScriptGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/script/config/test1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "test1",
			"alias": "Test Script",
			"description": "Test description",
			"mode": "single",
			"sequence": [{"service": "light.turn_on"}],
			"variables": {"test_var": "test_value"}
		}`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	config, err := client.ScriptGet(ctx, "test1")
	if err != nil {
		t.Fatalf("ScriptGet() error = %v", err)
	}

	if config.ID != "test1" {
		t.Errorf("expected id test1, got %s", config.ID)
	}
	if config.Alias != "Test Script" {
		t.Errorf("expected alias 'Test Script', got %s", config.Alias)
	}
	if config.Description == nil || *config.Description != "Test description" {
		t.Error("expected description 'Test description'")
	}
	if config.Mode != "single" {
		t.Errorf("expected mode 'single', got %s", config.Mode)
	}
}

func TestClient_ScriptGet_NoID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	_, err := client.ScriptGet(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestClient_ScriptSave(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/script/config/test1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var config map[string]any
		json.NewDecoder(r.Body).Decode(&config)

		if config["id"] != "test1" {
			t.Errorf("expected id test1, got %v", config["id"])
		}
		if config["alias"] != "Test Script" {
			t.Errorf("expected alias 'Test Script', got %v", config["alias"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	desc := "Test description"
	config := &ScriptConfig{
		ID:          "test1",
		Alias:       "Test Script",
		Description: &desc,
		Mode:        "single",
		Sequence:    []any{map[string]any{"service": "light.turn_on"}},
	}

	err := client.ScriptSave(ctx, config)
	if err != nil {
		t.Fatalf("ScriptSave() error = %v", err)
	}
}

func TestClient_ScriptSave_NilConfig(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptSave(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestClient_ScriptSave_NoID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	config := &ScriptConfig{
		Alias:    "Test",
		Sequence: []any{},
	}

	err := client.ScriptSave(ctx, config)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestClient_ScriptSave_NoAlias(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	config := &ScriptConfig{
		ID:       "test1",
		Sequence: []any{},
	}

	err := client.ScriptSave(ctx, config)
	if err == nil {
		t.Fatal("expected error for missing alias")
	}
}

func TestClient_ScriptSave_NoSequence(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	config := &ScriptConfig{
		ID:    "test1",
		Alias: "Test Script",
	}

	err := client.ScriptSave(ctx, config)
	if err == nil {
		t.Fatal("expected error for missing sequence")
	}
}

func TestClient_ScriptDeleteConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/script/config/test1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.ScriptDeleteConfig(ctx, "test1")
	if err != nil {
		t.Fatalf("ScriptDeleteConfig() error = %v", err)
	}
}

func TestClient_ScriptDeleteConfig_NoID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.ScriptDeleteConfig(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}
