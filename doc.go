// Package hago provides a Go client library for the Home Assistant REST API.
//
// The library provides full coverage of the Home Assistant REST API endpoints
// including state management, service calls, event handling, history, and more.
//
// # Authentication
//
// All requests require a Long-Lived Access Token from Home Assistant.
// Generate one from your Home Assistant profile page at:
// http://your-ha-instance:8123/profile -> Long-Lived Access Tokens
//
// # Basic Usage
//
//	client, err := hago.New(
//	    hago.WithBaseURL("http://homeassistant.local:8123"),
//	    hago.WithToken("your-long-lived-access-token"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Check API status
//	status, err := client.Status(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(status.Message)
//
//	// Get all entity states
//	states, err := client.States(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, state := range states {
//	    fmt.Printf("%s: %s\n", state.EntityID, state.State)
//	}
//
// # Service Calls
//
// Call Home Assistant services to control devices:
//
//	err := client.CallService(ctx, "light", "turn_on", &hago.ServiceData{
//	    EntityID: "light.living_room",
//	    Data: map[string]any{
//	        "brightness": 255,
//	    },
//	})
//
// # Thread Safety
//
// The Client is safe for concurrent use by multiple goroutines.
package hago
