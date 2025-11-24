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

The `hago` CLI provides a command-line interface for testing and interacting with Home Assistant. It uses [Cobra](https://github.com/spf13/cobra) for subcommands and [Viper](https://github.com/spf13/viper) for configuration.

### Configuration

Configuration can be provided via (in order of precedence):

1. Command-line flags
2. Environment variables (`HAGO_URL`, `HAGO_TOKEN`, etc.)
3. Config file (`~/.hago.yaml` or `./.hago.yaml`)
4. `.env` file in current directory

#### Environment Variables

```bash
export HAGO_URL="http://homeassistant.local:8123"
export HAGO_TOKEN="your-long-lived-access-token"
```

#### Config File (~/.hago.yaml)

```yaml
url: "http://homeassistant.local:8123"
token: "your-long-lived-access-token"
timeout: 30s
log_level: info
log_format: text
```

#### .env File

```bash
HAGO_URL=http://homeassistant.local:8123
HAGO_TOKEN=your-long-lived-access-token
```

### Commands

```bash
# Check API status
hago status

# Get Home Assistant configuration
hago config
hago config check  # Validate configuration.yaml

# List loaded components
hago components

# Entity states
hago state list                           # List all entities
hago state get light.living_room          # Get specific entity
hago state set sensor.test 42             # Set entity state
hago state set sensor.test 42 --attr '{"unit": "celsius"}'
hago state delete sensor.test             # Delete entity

# Services
hago service list                         # List all services
hago service call light turn_on light.living_room
hago service call light turn_on light.living_room -d '{"brightness": 255}'

# Events
hago event list                           # List event types
hago event fire my_event -d '{"key": "value"}'

# History
hago history light.living_room            # Last 24 hours
hago history light.living_room -d 48h     # Last 48 hours
hago history sensor.temp -d 7d --minimal  # 7 days, minimal response

# Logbook
hago logbook                              # Last 24 hours
hago logbook -d 12h                       # Last 12 hours
hago logbook -e light.living_room         # Filter by entity

# Error log
hago errorlog

# Templates
hago template "{{ states('light.living_room') }}"
hago template "{{ now().strftime('%Y-%m-%d') }}"

# Calendars
hago calendar list                        # List all calendars
hago calendar events calendar.personal    # Next 7 days
hago calendar events calendar.work -d 14  # Next 14 days

# Shell completion
hago completion bash > /etc/bash_completion.d/hago
hago completion zsh > "${fpath[1]}/_hago"
hago completion fish > ~/.config/fish/completions/hago.fish
```

### Global Flags

| Flag | Environment | Description |
|------|-------------|-------------|
| `--url` | `HAGO_URL` | Home Assistant URL |
| `--token` | `HAGO_TOKEN` | Long-Lived Access Token |
| `--timeout` | `HAGO_TIMEOUT` | Request timeout (default: 30s) |
| `--log-level` | `HAGO_LOG_LEVEL` | Log level: debug, info, warn, error |
| `--log-format` | `HAGO_LOG_FORMAT` | Log format: text, json |
| `--output`, `-o` | - | Output format: json, pretty |
| `--config` | - | Config file path |

## Lovelace Dashboard Management

The library includes WebSocket API support for Lovelace dashboard management, enabling dashboard-as-code workflows.

### Library Usage

```go
// List all dashboards
dashboards, err := client.LovelaceListDashboards(ctx)

// Get dashboard configuration
config, err := client.LovelaceGetConfig(ctx, nil, false)  // default dashboard
config, err := client.LovelaceGetConfig(ctx, ptr("map"), false)  // specific dashboard

// Save dashboard configuration
err := client.LovelaceSaveConfig(ctx, nil, myConfig)

// Create a new dashboard
dashboard, err := client.LovelaceCreateDashboard(ctx, &hago.CreateDashboardRequest{
    URLPath: "my-dashboard",
    Title:   "My Dashboard",
    Icon:    "mdi:view-dashboard",
})

// List custom resources (cards, themes)
resources, err := client.LovelaceListResources(ctx)

// Close WebSocket when done (optional - auto-closes on program exit)
client.CloseWebSocket()
```

### CLI Usage

```bash
# List all dashboards
hago lovelace list

# Get dashboard configuration
hago lovelace get                    # default dashboard
hago lovelace get map                # specific dashboard
hago lovelace get --yaml > dash.yaml # export as YAML

# Save dashboard configuration
hago lovelace save -f dashboard.yaml
hago lovelace save map -f map.json
cat config.json | hago lovelace save

# Export all dashboards
hago lovelace export -o ./dashboards
hago lovelace export -o ./dashboards --yaml

# Create a new dashboard
hago lovelace create my-dash --title "My Dashboard" --icon mdi:home

# Delete dashboard configuration (reset to auto-gen)
hago lovelace delete map

# Remove dashboard entirely
hago lovelace remove-dashboard my-dash

# List custom resources
hago lovelace resources
```

## Features

- Full Home Assistant REST API coverage
- WebSocket API for Lovelace dashboard management
- Functional options pattern for configuration
- Context support for cancellation and timeouts
- Strongly typed requests and responses
- Thread-safe client
- CLI with Cobra/Viper for easy configuration
- Multiple config sources: flags, env vars, config files, .env

## API Coverage

### REST API
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

### WebSocket API
- [x] Lovelace dashboard list (`lovelace/dashboards/list`)
- [x] Lovelace config get (`lovelace/config`)
- [x] Lovelace config save (`lovelace/config/save`)
- [x] Lovelace config delete (`lovelace/config/delete`)
- [x] Lovelace dashboard create (`lovelace/dashboards/create`)
- [x] Lovelace dashboard update (`lovelace/dashboards/update`)
- [x] Lovelace dashboard delete (`lovelace/dashboards/delete`)
- [x] Lovelace resources list (`lovelace/resources`)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
