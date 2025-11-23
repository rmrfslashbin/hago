# hago

A Go client library for the Home Assistant REST API.

## Author

Robert Sigler

## License

MIT License - see [LICENSE](LICENSE) for details.

## Installation

### Library

```bash
go get github.com/rmrfslashbin/hago
```

### CLI

```bash
go install github.com/rmrfslashbin/hago/cmd/hago@latest
```

Or build from source:

```bash
git clone https://github.com/rmrfslashbin/hago.git
cd hago
make build
```

## Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rmrfslashbin/hago"
)

func main() {
    ctx := context.Background()

    client, err := hago.New(
        hago.WithBaseURL("http://homeassistant.local:8123"),
        hago.WithToken("your-long-lived-access-token"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Check API status
    status, err := client.Status(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(status.Message)

    // Get all entity states
    states, err := client.States(ctx)
    if err != nil {
        log.Fatal(err)
    }
    for _, state := range states {
        fmt.Printf("%s: %s\n", state.EntityID, state.State)
    }

    // Call a service
    _, err = client.CallService(ctx, "light", "turn_on", &hago.ServiceCallRequest{
        EntityID: "light.living_room",
        Data: map[string]any{
            "brightness": 255,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## CLI Usage

The `hago` CLI provides a command-line interface for testing and interacting with Home Assistant.

### Configuration

Set your Home Assistant URL and token via environment variables or flags:

```bash
export HAGO_URL="http://homeassistant.local:8123"
export HAGO_TOKEN="your-long-lived-access-token"
```

Or use flags:

```bash
hago -url "http://homeassistant.local:8123" -token "your-token" status
```

### Commands

```bash
# Check API status
hago status

# Get Home Assistant configuration
hago config

# List loaded components
hago components

# List all entity states
hago states

# Get state of a specific entity
hago state light.living_room

# List available services
hago services

# Call a service
hago call light turn_on light.living_room
hago call light turn_on light.living_room '{"brightness": 255}'

# List event types
hago events

# Fire an event
hago fire custom_event '{"key": "value"}'

# Get state history (default: last 24 hours)
hago history light.living_room
hago history light.living_room 48h

# Get logbook entries
hago logbook
hago logbook 12h

# Get error log
hago errorlog

# Render a Jinja2 template
hago template "{{ states('light.living_room') }}"

# List calendars
hago calendars

# Get calendar events (default: next 7 days)
hago calendar calendar.personal
hago calendar calendar.personal 14
```

### Flags

| Flag | Environment | Description |
|------|-------------|-------------|
| `-url` | `HAGO_URL` | Home Assistant URL |
| `-token` | `HAGO_TOKEN` | Long-Lived Access Token |
| `-timeout` | - | Request timeout (default: 30s) |
| `-log-level` | `HAGO_LOG_LEVEL` | Log level: debug, info, warn, error |
| `-log-format` | `HAGO_LOG_FORMAT` | Log format: text, json |
| `-version` | - | Print version information |

## Features

- Full Home Assistant REST API coverage
- Functional options pattern for configuration
- Context support for cancellation and timeouts
- Strongly typed requests and responses
- Thread-safe client
- Reference CLI for testing

## API Coverage

- [x] Core endpoints (`/api/`, `/api/config`, `/api/components`)
- [x] State management (`/api/states` - GET, POST, DELETE)
- [x] Service calls (`/api/services`)
- [x] Event handling (`/api/events`)
- [x] History and logbook (`/api/history`, `/api/logbook`, `/api/error_log`)
- [x] Camera proxy (`/api/camera_proxy`)
- [x] Calendar endpoints (`/api/calendars`)
- [x] Template rendering (`/api/template`)
- [x] Configuration check (`/api/config/core/check_config`)
- [x] Intent handling (`/api/intent/handle`)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
