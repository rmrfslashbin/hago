package hago

import (
	"context"
	"testing"

	"github.com/gorilla/websocket"
)

func TestClient_EntityRegistry(t *testing.T) {
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

		if cmd["type"] != "config/entity_registry/list" {
			t.Errorf("expected config/entity_registry/list, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{
					"entity_id":       "light.living_room",
					"name":            "Living Room Light",
					"area_id":         "living_room",
					"device_id":       "device123",
					"labels":          []string{"smart", "lighting"},
					"icon":            "mdi:lightbulb",
					"disabled_by":     nil,
					"hidden_by":       nil,
					"has_entity_name": true,
					"platform":        "hue",
					"unique_id":       "hue-001",
				},
				{
					"entity_id":       "sensor.temperature",
					"name":            nil,
					"area_id":         nil,
					"device_id":       "device456",
					"labels":          []string{},
					"icon":            nil,
					"disabled_by":     nil,
					"hidden_by":       nil,
					"has_entity_name": false,
					"platform":        "mqtt",
					"unique_id":       "mqtt-temp-001",
				},
			},
		})
	})
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	defer client.CloseWebSocket()
	ctx := context.Background()

	entries, err := client.EntityRegistry(ctx)
	if err != nil {
		t.Fatalf("EntityRegistry() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].EntityID != "light.living_room" {
		t.Errorf("expected entity_id light.living_room, got %s", entries[0].EntityID)
	}
	if entries[0].Name == nil || *entries[0].Name != "Living Room Light" {
		t.Error("expected name to be 'Living Room Light'")
	}
	if entries[0].AreaID == nil || *entries[0].AreaID != "living_room" {
		t.Error("expected area_id to be 'living_room'")
	}
	if len(entries[0].Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(entries[0].Labels))
	}
}

func TestClient_DeviceRegistry(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		conn.ReadJSON(&cmd)

		if cmd["type"] != "config/device_registry/list" {
			t.Errorf("expected config/device_registry/list, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{
					"id":                "device123",
					"name":              "Philips Hue Bridge",
					"area_id":           "office",
					"config_entries":    []string{"config1"},
					"connections":       [][]string{{"mac", "00:11:22:33:44:55"}},
					"disabled_by":       nil,
					"identifiers":       [][]string{{"hue", "bridge001"}},
					"manufacturer":      "Philips",
					"model":             "BSB002",
					"name_by_user":      "Hue Bridge Office",
					"sw_version":        "1.2.3",
					"hw_version":        "2.1",
					"serial_number":     "ABC123",
					"via_device_id":     nil,
					"configuration_url": "http://192.168.1.100",
					"entry_type":        "service",
					"labels":            []string{"bridge"},
				},
			},
		})
	})
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	defer client.CloseWebSocket()
	ctx := context.Background()

	entries, err := client.DeviceRegistry(ctx)
	if err != nil {
		t.Fatalf("DeviceRegistry() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].ID != "device123" {
		t.Errorf("expected id device123, got %s", entries[0].ID)
	}
	if entries[0].Manufacturer == nil || *entries[0].Manufacturer != "Philips" {
		t.Error("expected manufacturer to be 'Philips'")
	}
	if entries[0].Model == nil || *entries[0].Model != "BSB002" {
		t.Error("expected model to be 'BSB002'")
	}
}

func TestClient_AreaRegistry(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		conn.ReadJSON(&cmd)

		if cmd["type"] != "config/area_registry/list" {
			t.Errorf("expected config/area_registry/list, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{
					"area_id":  "living_room",
					"name":     "Living Room",
					"floor_id": "ground_floor",
					"icon":     "mdi:sofa",
					"picture":  "/local/living_room.jpg",
					"aliases":  []string{"lounge", "family room"},
					"labels":   []string{"main"},
				},
			},
		})
	})
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	defer client.CloseWebSocket()
	ctx := context.Background()

	entries, err := client.AreaRegistry(ctx)
	if err != nil {
		t.Fatalf("AreaRegistry() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].AreaID != "living_room" {
		t.Errorf("expected area_id living_room, got %s", entries[0].AreaID)
	}
	if entries[0].Name != "Living Room" {
		t.Errorf("expected name 'Living Room', got %s", entries[0].Name)
	}
	if len(entries[0].Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(entries[0].Aliases))
	}
}

func TestClient_LabelRegistry(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		conn.ReadJSON(&cmd)

		if cmd["type"] != "config/label_registry/list" {
			t.Errorf("expected config/label_registry/list, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{
					"label_id":    "security",
					"name":        "Security",
					"icon":        "mdi:shield",
					"color":       "#FF0000",
					"description": "Security-related devices",
				},
			},
		})
	})
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	defer client.CloseWebSocket()
	ctx := context.Background()

	entries, err := client.LabelRegistry(ctx)
	if err != nil {
		t.Fatalf("LabelRegistry() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].LabelID != "security" {
		t.Errorf("expected label_id security, got %s", entries[0].LabelID)
	}
	if entries[0].Name != "Security" {
		t.Errorf("expected name 'Security', got %s", entries[0].Name)
	}
	if entries[0].Color == nil || *entries[0].Color != "#FF0000" {
		t.Error("expected color to be '#FF0000'")
	}
}

func TestClient_FloorRegistry(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		conn.ReadJSON(&cmd)

		if cmd["type"] != "config/floor_registry/list" {
			t.Errorf("expected config/floor_registry/list, got %v", cmd["type"])
		}

		// Send response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": true,
			"result": []map[string]any{
				{
					"floor_id": "ground_floor",
					"name":     "Ground Floor",
					"icon":     "mdi:home",
					"level":    0,
					"aliases":  []string{"first floor", "main floor"},
				},
				{
					"floor_id": "upstairs",
					"name":     "Upstairs",
					"icon":     "mdi:stairs-up",
					"level":    1,
					"aliases":  []string{"second floor"},
				},
			},
		})
	})
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("test-token"))
	defer client.CloseWebSocket()
	ctx := context.Background()

	entries, err := client.FloorRegistry(ctx)
	if err != nil {
		t.Fatalf("FloorRegistry() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].FloorID != "ground_floor" {
		t.Errorf("expected floor_id ground_floor, got %s", entries[0].FloorID)
	}
	if entries[0].Level == nil || *entries[0].Level != 0 {
		t.Error("expected level to be 0")
	}
	if len(entries[0].Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(entries[0].Aliases))
	}
}

func TestClient_EntityRegistry_Error(t *testing.T) {
	server := mockWSServer(t, func(conn *websocket.Conn) {
		// Auth flow
		conn.WriteJSON(map[string]any{"type": "auth_required"})
		var auth map[string]any
		conn.ReadJSON(&auth)
		conn.WriteJSON(map[string]any{"type": "auth_ok"})

		// Read command
		var cmd map[string]any
		conn.ReadJSON(&cmd)

		// Send error response
		conn.WriteJSON(map[string]any{
			"id":      cmd["id"],
			"type":    "result",
			"success": false,
			"error": map[string]any{
				"code":    "unauthorized",
				"message": "Unauthorized",
			},
		})
	})
	defer server.Close()

	client, _ := New(WithBaseURL(server.URL), WithToken("bad-token"))
	defer client.CloseWebSocket()
	ctx := context.Background()

	_, err := client.EntityRegistry(ctx)
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}
