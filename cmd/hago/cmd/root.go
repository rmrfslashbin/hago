// Package cmd implements the CLI commands for hago.
package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Build information set via ldflags.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var (
	cfgFile string
	client  *hago.Client
	logger  *slog.Logger
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "hago",
	Short: "Home Assistant Go CLI",
	Long: `hago is a command-line interface for the Home Assistant REST API.

It allows you to interact with your Home Assistant instance from the terminal,
including querying states, calling services, and more.

Configuration can be provided via:
  - Command-line flags
  - Environment variables (HAGO_URL, HAGO_TOKEN, etc.)
  - Config file (~/.hago.yaml or ./.hago.yaml)
  - .env file in current directory`,
	PersistentPreRunE: initializeClient,
	SilenceUsage:      true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.hago.yaml)")
	rootCmd.PersistentFlags().String("url", "", "Home Assistant URL")
	rootCmd.PersistentFlags().String("token", "", "Long-Lived Access Token")
	rootCmd.PersistentFlags().Duration("timeout", 30*time.Second, "Request timeout")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json)")

	// Bind flags to viper
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hago %s\n", Version)
			fmt.Printf("  commit:  %s\n", GitCommit)
			fmt.Printf("  built:   %s\n", BuildTime)
		},
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Load .env file if present
	if err := godotenv.Load(); err == nil {
		// .env loaded successfully
	}

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err == nil {
			// Search for config in home directory
			viper.AddConfigPath(home)
		}
		// Also search in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName(".hago")
		viper.SetConfigType("yaml")
	}

	// Environment variables
	viper.SetEnvPrefix("HAGO")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	// Read config file if present
	if err := viper.ReadInConfig(); err == nil {
		// Config file found and read
	}
}

// initializeClient creates the Home Assistant client.
func initializeClient(cmd *cobra.Command, args []string) error {
	// Skip client initialization for version command
	if cmd.Name() == "version" || cmd.Name() == "help" || cmd.Name() == "completion" {
		return nil
	}

	// Setup logger
	var err error
	logger, err = setupLogger(viper.GetString("log_level"), viper.GetString("log_format"))
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	slog.SetDefault(logger)

	// Get configuration
	url := viper.GetString("url")
	token := viper.GetString("token")
	timeout := viper.GetDuration("timeout")

	if url == "" {
		return fmt.Errorf("home Assistant URL is required (use --url, HAGO_URL, or config file)")
	}
	if token == "" {
		return fmt.Errorf("access token is required (use --token, HAGO_TOKEN, or config file)")
	}

	// Create client
	client, err = hago.New(
		hago.WithBaseURL(url),
		hago.WithToken(token),
		hago.WithTimeout(timeout),
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	logger.Debug("client initialized",
		"url", url,
		"timeout", timeout,
	)

	return nil
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

// getClient returns the initialized client.
func getClient() *hago.Client {
	return client
}
