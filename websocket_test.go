package hago

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockWSServer creates a mock WebSocket server for testing.
func mockWSServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websocket" {
			http.NotFound(w, r)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
			return
		}
		defer conn.Close()

		handler(conn)
	}))

	return server
}

func TestClient_WebSocketConnect(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Send auth_required
		conn.WriteJSON(map[string]any{
			"type":       "auth_required",
			"ha_version": "2024.1.0",
		})

		// Read auth message
		var auth map[string]any
		if err := conn.ReadJSON(&auth); err != nil {
			t.Errorf("read auth: %v", err)
			return
		}

		if auth["type"] != "auth" {
			t.Errorf("expected auth type, got %v", auth["type"])
		}
		if auth["access_token"] != "test-token" {
			t.Errorf("expected test-token, got %v", auth["access_token"])
		}

		// Send auth_ok
		conn.WriteJSON(map[string]any{
			"type":       "auth_ok",
			"ha_version": "2024.1.0",
		})

		// Keep connection open briefly
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Convert HTTP URL to WS URL for our client
	httpURL := server.URL

	client, err := New(
		WithBaseURL(httpURL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.CloseWebSocket()

	ctx := context.Background()
	if err := client.connectWebSocket(ctx); err != nil {
		t.Fatalf("connectWebSocket() error = %v", err)
	}
}

func TestClient_WebSocketAuthFailed(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Send auth_required
		conn.WriteJSON(map[string]any{
			"type":       "auth_required",
			"ha_version": "2024.1.0",
		})

		// Read auth message
		var auth map[string]any
		conn.ReadJSON(&auth)

		// Send auth_invalid
		conn.WriteJSON(map[string]any{
			"type":    "auth_invalid",
			"message": "Invalid access token",
		})
	})
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("bad-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()
	err = client.connectWebSocket(ctx)
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !strings.Contains(err.Error(), "auth failed") {
		t.Errorf("expected auth failed error, got: %v", err)
	}
}

func TestClient_LovelaceListDashboards(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		if err := conn.ReadJSON(&cmd); err != nil {
			t.Errorf("read command: %v", err)
			return
		}

		if cmd["type"] != "lovelace/dashboards/list" {
			t.Errorf("expected lovelace/dashboards/list, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{"url_path": "", "title": "Overview", "mode": "storage"},
				{"url_path": "map", "title": "Map", "mode": "storage"},
			},
		})
	})
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.CloseWebSocket()

	ctx := context.Background()
	dashboards, err := client.LovelaceListDashboards(ctx)
	if err != nil {
		t.Fatalf("LovelaceListDashboards() error = %v", err)
	}

	if len(dashboards) != 2 {
		t.Errorf("expected 2 dashboards, got %d", len(dashboards))
	}
}

func TestClient_LovelaceGetConfig(t *testing.T) {
	expectedConfig := map[string]any{
		"title": "Home",
		"views": []any{
			map[string]any{"title": "Overview", "path": "overview"},
		},
	}

	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		if err := conn.ReadJSON(&cmd); err != nil {
			t.Errorf("read command: %v", err)
			return
		}

		if cmd["type"] != "lovelace/config" {
			t.Errorf("expected lovelace/config, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result":  expectedConfig,
		})
	})
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.CloseWebSocket()

	ctx := context.Background()
	config, err := client.LovelaceGetConfig(ctx, nil, false)
	if err != nil {
		t.Fatalf("LovelaceGetConfig() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(config, &parsed); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}

	if parsed["title"] != "Home" {
		t.Errorf("expected title 'Home', got %v", parsed["title"])
	}
}

func TestClient_LovelaceSaveConfig(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		if err := conn.ReadJSON(&cmd); err != nil {
			t.Errorf("read command: %v", err)
			return
		}

		if cmd["type"] != "lovelace/config/save" {
			t.Errorf("expected lovelace/config/save, got %v", cmd["type"])
		}

		// Verify config was sent
		if cmd["config"] == nil {
			t.Error("expected config in command")
		}

		// Send success response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result":  nil,
		})
	})
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.CloseWebSocket()

	ctx := context.Background()
	config := map[string]any{
		"title": "Test",
		"views": []any{},
	}

	if err := client.LovelaceSaveConfig(ctx, nil, config); err != nil {
		t.Fatalf("LovelaceSaveConfig() error = %v", err)
	}
}

func TestClient_LovelaceListResources(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		if err := conn.ReadJSON(&cmd); err != nil {
			t.Errorf("read command: %v", err)
			return
		}

		if cmd["type"] != "lovelace/resources" {
			t.Errorf("expected lovelace/resources, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{"id": "1", "type": "module", "url": "/local/card.js"},
			},
		})
	})
	defer server.Close()

	client, err := New(
		WithBaseURL(server.URL),
		WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.CloseWebSocket()

	ctx := context.Background()
	resources, err := client.LovelaceListResources(ctx)
	if err != nil {
		t.Fatalf("LovelaceListResources() error = %v", err)
	}

	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
}

func TestWebSocketError_Error(t *testing.T) {
	err := &WebSocketError{
		Code:    "config_not_found",
		Message: "Dashboard not found",
	}

	expected := "websocket error [config_not_found]: Dashboard not found"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestClient_BuildWebSocketURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "http",
			baseURL:  "http://localhost:8123",
			expected: "ws://localhost:8123/api/websocket",
		},
		{
			name:     "https",
			baseURL:  "https://ha.example.com",
			expected: "wss://ha.example.com/api/websocket",
		},
		{
			name:     "with trailing slash",
			baseURL:  "http://localhost:8123/",
			expected: "ws://localhost:8123/api/websocket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := New(
				WithBaseURL(tt.baseURL),
				WithToken("test"),
			)

			url, err := client.buildWebSocketURL()
			if err != nil {
				t.Fatalf("buildWebSocketURL() error = %v", err)
			}

			if url != tt.expected {
				t.Errorf("buildWebSocketURL() = %v, want %v", url, tt.expected)
			}
		})
	}
}
