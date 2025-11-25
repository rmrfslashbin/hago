package hago

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_AutomationTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/automation/trigger" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req["entity_id"] != "automation.test" {
			t.Errorf("expected entity_id automation.test, got %v", req["entity_id"])
		}

		// skip_condition is flattened to top level by ServiceCallRequest.MarshalJSON
		if req["skip_condition"] != true {
			t.Errorf("expected skip_condition to be true, got %v", req["skip_condition"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	skipCond := true
	err := client.AutomationTrigger(ctx, &AutomationTriggerRequest{
		EntityID:      "automation.test",
		SkipCondition: &skipCond,
	})
	if err != nil {
		t.Fatalf("AutomationTrigger() error = %v", err)
	}
}

func TestClient_AutomationTrigger_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.AutomationTrigger(ctx, &AutomationTriggerRequest{})
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_AutomationTurnOn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/automation/turn_on" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "automation.test" {
			t.Errorf("expected entity_id automation.test, got %v", req["entity_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.AutomationTurnOn(ctx, "automation.test")
	if err != nil {
		t.Fatalf("AutomationTurnOn() error = %v", err)
	}
}

func TestClient_AutomationTurnOn_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.AutomationTurnOn(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_AutomationTurnOff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/automation/turn_off" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "automation.test" {
			t.Errorf("expected entity_id automation.test, got %v", req["entity_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.AutomationTurnOff(ctx, "automation.test")
	if err != nil {
		t.Fatalf("AutomationTurnOff() error = %v", err)
	}
}

func TestClient_AutomationTurnOff_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.AutomationTurnOff(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_AutomationToggle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/automation/toggle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["entity_id"] != "automation.test" {
			t.Errorf("expected entity_id automation.test, got %v", req["entity_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.AutomationToggle(ctx, "automation.test")
	if err != nil {
		t.Fatalf("AutomationToggle() error = %v", err)
	}
}

func TestClient_AutomationToggle_NoEntityID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.AutomationToggle(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing entity_id")
	}
}

func TestClient_AutomationReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/automation/reload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.AutomationReload(ctx)
	if err != nil {
		t.Fatalf("AutomationReload() error = %v", err)
	}
}

func TestClient_AutomationTrigger_WithoutSkipCondition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		// skip_condition should not be present when not set
		if _, ok := req["skip_condition"]; ok {
			t.Error("expected no skip_condition field when not provided")
		}

		if req["entity_id"] != "automation.test" {
			t.Errorf("expected entity_id automation.test, got %v", req["entity_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	err := client.AutomationTrigger(ctx, &AutomationTriggerRequest{
		EntityID: "automation.test",
		// SkipCondition not set
	})
	if err != nil {
		t.Fatalf("AutomationTrigger() error = %v", err)
	}
}

func TestClient_AutomationList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/automation/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": "test1",
				"alias": "Test Automation 1",
				"trigger": [{"platform": "state"}],
				"action": [{"service": "light.turn_on"}]
			},
			{
				"id": "test2",
				"alias": "Test Automation 2",
				"description": "Test description",
				"mode": "single",
				"trigger": [{"platform": "time"}],
				"condition": [{"condition": "state"}],
				"action": [{"service": "light.turn_off"}]
			}
		]`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	configs, err := client.AutomationList(ctx)
	if err != nil {
		t.Fatalf("AutomationList() error = %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	if configs[0].ID != "test1" {
		t.Errorf("expected id test1, got %s", configs[0].ID)
	}
	if configs[0].Alias != "Test Automation 1" {
		t.Errorf("expected alias 'Test Automation 1', got %s", configs[0].Alias)
	}
	if configs[1].Description == nil || *configs[1].Description != "Test description" {
		t.Error("expected description 'Test description'")
	}
}

func TestClient_AutomationGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/automation/config/test1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "test1",
			"alias": "Test Automation",
			"description": "Test description",
			"mode": "single",
			"trigger": [{"platform": "state"}],
			"condition": [{"condition": "state"}],
			"action": [{"service": "light.turn_on"}]
		}`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	config, err := client.AutomationGet(ctx, "test1")
	if err != nil {
		t.Fatalf("AutomationGet() error = %v", err)
	}

	if config.ID != "test1" {
		t.Errorf("expected id test1, got %s", config.ID)
	}
	if config.Alias != "Test Automation" {
		t.Errorf("expected alias 'Test Automation', got %s", config.Alias)
	}
	if config.Description == nil || *config.Description != "Test description" {
		t.Error("expected description 'Test description'")
	}
	if config.Mode != "single" {
		t.Errorf("expected mode 'single', got %s", config.Mode)
	}
}

func TestClient_AutomationGet_NoID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	_, err := client.AutomationGet(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestClient_AutomationSave(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/automation/config/test1" {
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
		if config["alias"] != "Test Automation" {
			t.Errorf("expected alias 'Test Automation', got %v", config["alias"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	ctx := context.Background()

	desc := "Test description"
	config := &AutomationConfig{
		ID:          "test1",
		Alias:       "Test Automation",
		Description: &desc,
		Mode:        "single",
		Trigger:     []any{map[string]any{"platform": "state"}},
		Action:      []any{map[string]any{"service": "light.turn_on"}},
	}

	err := client.AutomationSave(ctx, config)
	if err != nil {
		t.Fatalf("AutomationSave() error = %v", err)
	}
}

func TestClient_AutomationSave_NilConfig(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.AutomationSave(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestClient_AutomationSave_NoID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	config := &AutomationConfig{
		Alias:   "Test",
		Trigger: []any{},
		Action:  []any{},
	}

	err := client.AutomationSave(ctx, config)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestClient_AutomationSave_NoAlias(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	config := &AutomationConfig{
		ID:      "test1",
		Trigger: []any{},
		Action:  []any{},
	}

	err := client.AutomationSave(ctx, config)
	if err == nil {
		t.Fatal("expected error for missing alias")
	}
}

func TestClient_AutomationDeleteConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config/automation/config/test1" {
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

	err := client.AutomationDeleteConfig(ctx, "test1")
	if err != nil {
		t.Fatalf("AutomationDeleteConfig() error = %v", err)
	}
}

func TestClient_AutomationDeleteConfig_NoID(t *testing.T) {
	client, _ := New(WithBaseURL("http://test"), WithToken("test"))
	ctx := context.Background()

	err := client.AutomationDeleteConfig(ctx, "")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}
