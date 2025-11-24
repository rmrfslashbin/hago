package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var lovelaceCmd = &cobra.Command{
	Use:   "lovelace",
	Short: "Manage Lovelace dashboards",
	Long: `Manage Lovelace dashboards via the WebSocket API.

This allows you to list, get, save, and delete dashboard configurations,
enabling dashboard-as-code workflows.`,
}

var lovelaceListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all dashboards",
	Long:    `List all Lovelace dashboards including storage-mode and YAML-mode dashboards.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		dashboards, err := getClient().LovelaceListDashboards(ctx)
		if err != nil {
			return err
		}
		return printResult(dashboards)
	},
}

var lovelaceGetCmd = &cobra.Command{
	Use:   "get [dashboard]",
	Short: "Get dashboard configuration",
	Long: `Get the configuration of a Lovelace dashboard.

If no dashboard is specified, returns the default (overview) dashboard.

Examples:
  hago lovelace get
  hago lovelace get map
  hago lovelace get --dashboard map
  hago lovelace get -o yaml > dashboard.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var urlPath *string
		dashboard, _ := cmd.Flags().GetString("dashboard")
		if dashboard != "" {
			urlPath = &dashboard
		} else if len(args) > 0 {
			urlPath = &args[0]
		}

		force, _ := cmd.Flags().GetBool("force")
		asYAML, _ := cmd.Flags().GetBool("yaml")

		config, err := getClient().LovelaceGetConfig(ctx, urlPath, force)
		if err != nil {
			return err
		}

		if asYAML {
			// Convert JSON to YAML
			var data any
			if err := json.Unmarshal(config, &data); err != nil {
				return err
			}
			return yaml.NewEncoder(os.Stdout).Encode(data)
		}

		return printResult(json.RawMessage(config))
	},
}

var lovelaceSaveCmd = &cobra.Command{
	Use:   "save [dashboard]",
	Short: "Save dashboard configuration",
	Long: `Save a Lovelace dashboard configuration.

The configuration can be provided via:
  - File (--file or -f)
  - Stdin (pipe or redirect)

Supports both JSON and YAML formats.

Examples:
  hago lovelace save -f dashboard.yaml
  hago lovelace save map -f map-dashboard.json
  cat dashboard.json | hago lovelace save
  hago lovelace get | jq '.title = "New Title"' | hago lovelace save`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var urlPath *string
		dashboard, _ := cmd.Flags().GetString("dashboard")
		if dashboard != "" {
			urlPath = &dashboard
		} else if len(args) > 0 {
			urlPath = &args[0]
		}

		// Read config from file or stdin
		file, _ := cmd.Flags().GetString("file")
		var data []byte
		var err error

		if file != "" {
			data, err = os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
		} else {
			// Read from stdin
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
		}

		if len(data) == 0 {
			return fmt.Errorf("no configuration provided (use --file or pipe to stdin)")
		}

		// Parse as JSON or YAML
		var config any
		if err := json.Unmarshal(data, &config); err != nil {
			// Try YAML
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("parse config (tried JSON and YAML): %w", err)
			}
		}

		if err := getClient().LovelaceSaveConfig(ctx, urlPath, config); err != nil {
			return err
		}

		name := "default"
		if urlPath != nil {
			name = *urlPath
		}
		printSuccess("Dashboard '%s' saved successfully", name)
		return nil
	},
}

var lovelaceDeleteCmd = &cobra.Command{
	Use:     "delete [dashboard]",
	Aliases: []string{"rm"},
	Short:   "Delete dashboard configuration",
	Long: `Delete a Lovelace dashboard configuration.

This resets the dashboard to auto-generated mode (not the same as deleting the dashboard itself).

Examples:
  hago lovelace delete
  hago lovelace delete map`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var urlPath *string
		dashboard, _ := cmd.Flags().GetString("dashboard")
		if dashboard != "" {
			urlPath = &dashboard
		} else if len(args) > 0 {
			urlPath = &args[0]
		}

		if err := getClient().LovelaceDeleteConfig(ctx, urlPath); err != nil {
			return err
		}

		name := "default"
		if urlPath != nil {
			name = *urlPath
		}
		printSuccess("Dashboard '%s' configuration deleted (reset to auto-gen)", name)
		return nil
	},
}

var lovelaceCreateCmd = &cobra.Command{
	Use:   "create <url_path>",
	Short: "Create a new dashboard",
	Long: `Create a new Lovelace dashboard.

Examples:
  hago lovelace create my-dashboard --title "My Dashboard"
  hago lovelace create admin-panel --title "Admin" --require-admin --icon mdi:shield`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		req := &hago.CreateDashboardRequest{
			URLPath: args[0],
		}

		if title, _ := cmd.Flags().GetString("title"); title != "" {
			req.Title = title
		}
		if icon, _ := cmd.Flags().GetString("icon"); icon != "" {
			req.Icon = icon
		}
		req.ShowInSidebar, _ = cmd.Flags().GetBool("sidebar")
		req.RequireAdmin, _ = cmd.Flags().GetBool("require-admin")

		dashboard, err := getClient().LovelaceCreateDashboard(ctx, req)
		if err != nil {
			return err
		}

		return printResult(dashboard)
	},
}

var lovelaceRemoveDashboardCmd = &cobra.Command{
	Use:   "remove-dashboard <dashboard_id>",
	Short: "Remove a dashboard entirely",
	Long: `Remove a dashboard entirely (not just its configuration).

This is different from 'delete' which only resets the config to auto-gen mode.

Examples:
  hago lovelace remove-dashboard my-dashboard`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().LovelaceDeleteDashboard(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Dashboard '%s' removed", args[0])
		return nil
	},
}

var lovelaceResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "List Lovelace resources",
	Long:  `List all registered Lovelace resources (custom cards, themes, etc).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		resources, err := getClient().LovelaceListResources(ctx)
		if err != nil {
			return err
		}
		return printResult(resources)
	},
}

var lovelaceExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all dashboard configurations",
	Long: `Export all Lovelace dashboard configurations to files.

Creates one file per dashboard in the output directory.

Examples:
  hago lovelace export -o ./dashboards
  hago lovelace export -o ./dashboards --yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		outDir, _ := cmd.Flags().GetString("output")
		asYAML, _ := cmd.Flags().GetBool("yaml")

		if outDir == "" {
			outDir = "."
		}

		// Create output directory
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}

		// List dashboards
		dashboards, err := getClient().LovelaceListDashboards(ctx)
		if err != nil {
			return err
		}

		exported := 0
		for _, dash := range dashboards {
			urlPath := dash.URLPath
			var pathPtr *string
			if urlPath != "" {
				pathPtr = &urlPath
			}

			config, err := getClient().LovelaceGetConfig(ctx, pathPtr, false)
			if err != nil {
				getLogger().Warn("failed to get config", "dashboard", urlPath, "error", err)
				continue
			}

			// Determine filename
			name := urlPath
			if name == "" {
				name = "default"
			}

			var filename string
			var data []byte

			if asYAML {
				filename = fmt.Sprintf("%s/%s.yaml", outDir, name)
				var parsed any
				if err := json.Unmarshal(config, &parsed); err != nil {
					getLogger().Warn("failed to parse config", "dashboard", urlPath, "error", err)
					continue
				}
				data, err = yaml.Marshal(parsed)
				if err != nil {
					getLogger().Warn("failed to marshal YAML", "dashboard", urlPath, "error", err)
					continue
				}
			} else {
				filename = fmt.Sprintf("%s/%s.json", outDir, name)
				// Pretty print JSON
				var parsed any
				json.Unmarshal(config, &parsed)
				data, _ = json.MarshalIndent(parsed, "", "  ")
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				getLogger().Warn("failed to write file", "file", filename, "error", err)
				continue
			}

			exported++
			printSuccess("Exported: %s", filename)
		}

		printSuccess("\nExported %d dashboard(s)", exported)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lovelaceCmd)

	lovelaceCmd.AddCommand(lovelaceListCmd)
	lovelaceCmd.AddCommand(lovelaceGetCmd)
	lovelaceCmd.AddCommand(lovelaceSaveCmd)
	lovelaceCmd.AddCommand(lovelaceDeleteCmd)
	lovelaceCmd.AddCommand(lovelaceCreateCmd)
	lovelaceCmd.AddCommand(lovelaceRemoveDashboardCmd)
	lovelaceCmd.AddCommand(lovelaceResourcesCmd)
	lovelaceCmd.AddCommand(lovelaceExportCmd)

	// Get flags
	lovelaceGetCmd.Flags().StringP("dashboard", "d", "", "Dashboard URL path")
	lovelaceGetCmd.Flags().Bool("force", false, "Bypass cache")
	lovelaceGetCmd.Flags().Bool("yaml", false, "Output as YAML")

	// Save flags
	lovelaceSaveCmd.Flags().StringP("dashboard", "d", "", "Dashboard URL path")
	lovelaceSaveCmd.Flags().StringP("file", "f", "", "Config file (JSON or YAML)")

	// Delete flags
	lovelaceDeleteCmd.Flags().StringP("dashboard", "d", "", "Dashboard URL path")

	// Create flags
	lovelaceCreateCmd.Flags().StringP("title", "t", "", "Dashboard title")
	lovelaceCreateCmd.Flags().StringP("icon", "i", "", "Dashboard icon (e.g., mdi:home)")
	lovelaceCreateCmd.Flags().Bool("sidebar", true, "Show in sidebar")
	lovelaceCreateCmd.Flags().Bool("require-admin", false, "Require admin access")

	// Export flags
	lovelaceExportCmd.Flags().StringP("output", "o", ".", "Output directory")
	lovelaceExportCmd.Flags().Bool("yaml", false, "Export as YAML")
}
