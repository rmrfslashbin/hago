# hago

A Go client library for the Home Assistant REST API.

## Author

Robert Sigler

## License

MIT License - see [LICENSE](LICENSE) for details.

## Installation

```bash
go get github.com/rmrfslashbin/hago
```

## Usage

```go
package main

import (
    "github.com/rmrfslashbin/hago"
)

func main() {
    client := hago.NewClient("http://homeassistant.local:8123", "your-long-lived-access-token")

    // Check API status
    status, err := client.Ping()
    if err != nil {
        panic(err)
    }

    // Get all entity states
    states, err := client.States()
    if err != nil {
        panic(err)
    }
}
```

## Features

- Full Home Assistant REST API coverage
- Simple, idiomatic Go interface
- Bearer token authentication
- Strongly typed responses

## API Coverage

- [ ] Core endpoints (`/api/`, `/api/config`, `/api/components`)
- [ ] State management (`/api/states`)
- [ ] Service calls (`/api/services`)
- [ ] Event handling (`/api/events`)
- [ ] History and logbook (`/api/history`, `/api/logbook`)
- [ ] Camera proxy (`/api/camera_proxy`)
- [ ] Calendar endpoints (`/api/calendars`)
- [ ] Template rendering (`/api/template`)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
