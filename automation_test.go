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
