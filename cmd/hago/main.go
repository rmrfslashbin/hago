// hago is a command-line interface for interacting with the Home Assistant REST API.
//
// Usage:
//
//	hago [flags] <command> [command-flags]
//
// Available commands:
//
//	status      Check API status
//	config      Get Home Assistant configuration
//	components  List loaded components
//	states      List or get entity states
//	state       Get or set a specific entity state
//	services    List available services
//	call        Call a service
//	events      List event types
//	fire        Fire an event
//	history     Get state history
//	logbook     Get logbook entries
//	errorlog    Get error log
//	template    Render a template
//	calendars   List calendars or get calendar events
//
// Global flags:
//
//	-url        Home Assistant URL (or HAGO_URL env var)
//	-token      Long-Lived Access Token (or HAGO_TOKEN env var)
//	-timeout    Request timeout (default: 30s)
//	-log-level  Log level: debug, info, warn, error (default: info)
//	-log-format Log format: text, json (default: text)
//	-version    Print version information
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rmrfslashbin/hago"
)

// Build information set via ldflags.
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

// Config holds CLI configuration.
type Config struct {
	URL       string
	Token     string
	Timeout   time.Duration
	LogLevel  string
	LogFormat string
}

func main() {
	os.Exit(run())
}

func run() int {
	// Global flags
	cfg := &Config{}
	flag.StringVar(&cfg.URL, "url", getEnv("HAGO_URL", ""), "Home Assistant URL")
	flag.StringVar(&cfg.Token, "token", getEnv("HAGO_TOKEN", ""), "Long-Lived Access Token")
	flag.DurationVar(&cfg.Timeout, "timeout", 30*time.Second, "Request timeout")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnv("HAGO_LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.LogFormat, "log-format", getEnv("HAGO_LOG_FORMAT", "text"), "Log format (text, json)")
	showVersion := flag.Bool("version", false, "Print version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "hago - Home Assistant Go CLI\n\n")
		fmt.Fprintf(os.Stderr, "Usage: hago [flags] <command> [args]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  status              Check API status\n")
		fmt.Fprintf(os.Stderr, "  config              Get Home Assistant configuration\n")
		fmt.Fprintf(os.Stderr, "  components          List loaded components\n")
		fmt.Fprintf(os.Stderr, "  states              List all entity states\n")
		fmt.Fprintf(os.Stderr, "  state <entity_id>   Get state of an entity\n")
		fmt.Fprintf(os.Stderr, "  services            List available services\n")
		fmt.Fprintf(os.Stderr, "  call <domain> <service> [entity_id] [json_data]\n")
		fmt.Fprintf(os.Stderr, "                      Call a service\n")
		fmt.Fprintf(os.Stderr, "  events              List event types\n")
		fmt.Fprintf(os.Stderr, "  fire <event_type> [json_data]\n")
		fmt.Fprintf(os.Stderr, "                      Fire an event\n")
		fmt.Fprintf(os.Stderr, "  history <entity_id> [duration]\n")
		fmt.Fprintf(os.Stderr, "                      Get state history (default: 24h)\n")
		fmt.Fprintf(os.Stderr, "  logbook [duration]  Get logbook entries (default: 24h)\n")
		fmt.Fprintf(os.Stderr, "  errorlog            Get error log\n")
		fmt.Fprintf(os.Stderr, "  template <template> Render a Jinja2 template\n")
		fmt.Fprintf(os.Stderr, "  calendars           List calendars\n")
		fmt.Fprintf(os.Stderr, "  calendar <entity_id> [days]\n")
		fmt.Fprintf(os.Stderr, "                      Get calendar events (default: 7 days)\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  HAGO_URL            Home Assistant URL\n")
		fmt.Fprintf(os.Stderr, "  HAGO_TOKEN          Long-Lived Access Token\n")
		fmt.Fprintf(os.Stderr, "  HAGO_LOG_LEVEL      Log level\n")
		fmt.Fprintf(os.Stderr, "  HAGO_LOG_FORMAT     Log format\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("hago %s\n", version)
		fmt.Printf("  commit:  %s\n", gitCommit)
		fmt.Printf("  built:   %s\n", buildTime)
		return 0
	}

	// Setup logger
	logger, err := setupLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	slog.SetDefault(logger)

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return 1
	}

	// Validate required config
	if cfg.URL == "" {
		fmt.Fprintf(os.Stderr, "Error: Home Assistant URL is required (use -url or HAGO_URL)\n")
		return 1
	}
	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "Error: Access token is required (use -token or HAGO_TOKEN)\n")
		return 1
	}

	// Create client
	client, err := hago.New(
		hago.WithBaseURL(cfg.URL),
		hago.WithToken(cfg.Token),
		hago.WithTimeout(cfg.Timeout),
	)
	if err != nil {
		logger.Error("failed to create client", "error", err)
		return 1
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Debug("received interrupt signal")
		cancel()
	}()

	// Execute command
	cmd := args[0]
	cmdArgs := args[1:]

	if err := executeCommand(ctx, client, cmd, cmdArgs); err != nil {
		logger.Error("command failed", "command", cmd, "error", err)
		return 1
	}

	return 0
}

func executeCommand(ctx context.Context, client *hago.Client, cmd string, args []string) error {
	switch cmd {
	case "status":
		return cmdStatus(ctx, client)
	case "config":
		return cmdConfig(ctx, client)
	case "components":
		return cmdComponents(ctx, client)
	case "states":
		return cmdStates(ctx, client)
	case "state":
		if len(args) < 1 {
			return fmt.Errorf("usage: hago state <entity_id>")
		}
		return cmdState(ctx, client, args[0])
	case "services":
		return cmdServices(ctx, client)
	case "call":
		if len(args) < 2 {
			return fmt.Errorf("usage: hago call <domain> <service> [entity_id] [json_data]")
		}
		var entityID, jsonData string
		if len(args) > 2 {
			entityID = args[2]
		}
		if len(args) > 3 {
			jsonData = args[3]
		}
		return cmdCall(ctx, client, args[0], args[1], entityID, jsonData)
	case "events":
		return cmdEvents(ctx, client)
	case "fire":
		if len(args) < 1 {
			return fmt.Errorf("usage: hago fire <event_type> [json_data]")
		}
		var jsonData string
		if len(args) > 1 {
			jsonData = args[1]
		}
		return cmdFire(ctx, client, args[0], jsonData)
	case "history":
		if len(args) < 1 {
			return fmt.Errorf("usage: hago history <entity_id> [duration]")
		}
		duration := "24h"
		if len(args) > 1 {
			duration = args[1]
		}
		return cmdHistory(ctx, client, args[0], duration)
	case "logbook":
		duration := "24h"
		if len(args) > 0 {
			duration = args[0]
		}
		return cmdLogbook(ctx, client, duration)
	case "errorlog":
		return cmdErrorLog(ctx, client)
	case "template":
		if len(args) < 1 {
			return fmt.Errorf("usage: hago template <template>")
		}
		return cmdTemplate(ctx, client, args[0])
	case "calendars":
		return cmdCalendars(ctx, client)
	case "calendar":
		if len(args) < 1 {
			return fmt.Errorf("usage: hago calendar <entity_id> [days]")
		}
		days := 7
		if len(args) > 1 {
			fmt.Sscanf(args[1], "%d", &days)
		}
		return cmdCalendar(ctx, client, args[0], days)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func cmdStatus(ctx context.Context, client *hago.Client) error {
	status, err := client.Status(ctx)
	if err != nil {
		return err
	}
	return printJSON(status)
}

func cmdConfig(ctx context.Context, client *hago.Client) error {
	config, err := client.Config(ctx)
	if err != nil {
		return err
	}
	return printJSON(config)
}

func cmdComponents(ctx context.Context, client *hago.Client) error {
	components, err := client.Components(ctx)
	if err != nil {
		return err
	}
	for _, c := range components {
		fmt.Println(c)
	}
	return nil
}

func cmdStates(ctx context.Context, client *hago.Client) error {
	states, err := client.States(ctx)
	if err != nil {
		return err
	}
	return printJSON(states)
}

func cmdState(ctx context.Context, client *hago.Client, entityID string) error {
	state, err := client.State(ctx, entityID)
	if err != nil {
		return err
	}
	return printJSON(state)
}

func cmdServices(ctx context.Context, client *hago.Client) error {
	services, err := client.Services(ctx)
	if err != nil {
		return err
	}
	return printJSON(services)
}

func cmdCall(ctx context.Context, client *hago.Client, domain, service, entityID, jsonData string) error {
	req := &hago.ServiceCallRequest{
		EntityID: entityID,
	}
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &req.Data); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
	}

	states, err := client.CallService(ctx, domain, service, req)
	if err != nil {
		return err
	}
	return printJSON(states)
}

func cmdEvents(ctx context.Context, client *hago.Client) error {
	events, err := client.Events(ctx)
	if err != nil {
		return err
	}
	return printJSON(events)
}

func cmdFire(ctx context.Context, client *hago.Client, eventType, jsonData string) error {
	var data hago.EventData
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
	}

	if err := client.FireEvent(ctx, eventType, data); err != nil {
		return err
	}
	fmt.Printf("Event '%s' fired successfully\n", eventType)
	return nil
}

func cmdHistory(ctx context.Context, client *hago.Client, entityID, durationStr string) error {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	startTime := time.Now().Add(-duration)
	opts := &hago.HistoryOptions{
		FilterEntityID: entityID,
	}

	history, err := client.History(ctx, startTime, opts)
	if err != nil {
		return err
	}
	return printJSON(history)
}

func cmdLogbook(ctx context.Context, client *hago.Client, durationStr string) error {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	startTime := time.Now().Add(-duration)
	entries, err := client.Logbook(ctx, startTime, nil)
	if err != nil {
		return err
	}
	return printJSON(entries)
}

func cmdErrorLog(ctx context.Context, client *hago.Client) error {
	log, err := client.ErrorLog(ctx)
	if err != nil {
		return err
	}
	fmt.Print(log)
	return nil
}

func cmdTemplate(ctx context.Context, client *hago.Client, template string) error {
	result, err := client.RenderTemplate(ctx, template)
	if err != nil {
		return err
	}
	fmt.Print(result)
	return nil
}

func cmdCalendars(ctx context.Context, client *hago.Client) error {
	calendars, err := client.Calendars(ctx)
	if err != nil {
		return err
	}
	return printJSON(calendars)
}

func cmdCalendar(ctx context.Context, client *hago.Client, entityID string, days int) error {
	start := time.Now()
	end := start.AddDate(0, 0, days)

	events, err := client.CalendarEvents(ctx, entityID, start, end)
	if err != nil {
		return err
	}
	return printJSON(events)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func setupLogger(level, format string) (*slog.Logger, error) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	var handler slog.Handler
	var writer io.Writer = os.Stderr

	if format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
