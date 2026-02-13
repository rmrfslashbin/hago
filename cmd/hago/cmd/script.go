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

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Manage Home Assistant scripts",
	Long: `Control scripts using the script service.

Provides commands to run, enable, disable, toggle, reload, and manage script configurations.`,
}

var scriptRunCmd = &cobra.Command{
	Use:   "run <entity_id>",
	Short: "Run a script",
	Long: `Run a script, optionally passing variables as JSON.

Examples:
  hago script run script.morning_routine
  hago script run script.notify_user --vars '{"message":"Hello","title":"Alert"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		varsJSON, _ := cmd.Flags().GetString("vars")

		var variables map[string]any
		if varsJSON != "" {
			if err := json.Unmarshal([]byte(varsJSON), &variables); err != nil {
				return fmt.Errorf("parse variables JSON: %w", err)
			}
		}

		if err := getClient().ScriptRun(ctx, args[0], variables); err != nil {
			return err
		}

		printSuccess("Script '%s' executed", args[0])
		return nil
	},
}

var scriptTurnOnCmd = &cobra.Command{
	Use:     "turn-on <entity_id>",
	Aliases: []string{"enable", "on"},
	Short:   "Turn on a script",
	Long: `Turn on (start) a script, optionally passing variables as JSON.

This is an alternative to 'run' that supports asynchronous execution.

Examples:
  hago script turn-on script.morning_routine
  hago script turn-on script.notify_user --vars '{"message":"Hello"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		varsJSON, _ := cmd.Flags().GetString("vars")

		var variables map[string]any
		if varsJSON != "" {
			if err := json.Unmarshal([]byte(varsJSON), &variables); err != nil {
				return fmt.Errorf("parse variables JSON: %w", err)
			}
		}

		if err := getClient().ScriptTurnOn(ctx, args[0], variables); err != nil {
			return err
		}

		printSuccess("Script '%s' turned on", args[0])
		return nil
	},
}

var scriptTurnOffCmd = &cobra.Command{
	Use:     "turn-off <entity_id>",
	Aliases: []string{"disable", "off", "stop"},
	Short:   "Turn off (stop) a running script",
	Long: `Turn off (stop) a running script.

Examples:
  hago script turn-off script.morning_routine
  hago script stop script.morning_routine`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().ScriptTurnOff(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Script '%s' turned off", args[0])
		return nil
	},
}

var scriptToggleCmd = &cobra.Command{
	Use:   "toggle <entity_id>",
	Short: "Toggle a script's running state",
	Long: `Toggle a script between running and stopped.

Examples:
  hago script toggle script.morning_routine`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().ScriptToggle(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Script '%s' toggled", args[0])
		return nil
	},
}

var scriptReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload all scripts from YAML",
	Long: `Reload all scripts from YAML configuration files.

This is useful after manually editing script YAML files to apply changes
without restarting Home Assistant.

Examples:
  hago script reload`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().ScriptReload(ctx); err != nil {
			return err
		}

		printSuccess("All scripts reloaded from YAML configuration")
		return nil
	},
}

var scriptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all script configurations",
	Long: `List all script configurations from Home Assistant.

WARNING: This uses an undocumented REST API endpoint that may change without notice.

Examples:
  hago script list
  hago script list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		configs, err := getClient().ScriptList(ctx)
		if err != nil {
			return err
		}

		return printResult(configs)
	},
}

var scriptGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get script configuration by ID",
	Long: `Get a specific script configuration by its ID.

WARNING: This uses an undocumented REST API endpoint that may change without notice.

Examples:
  hago script get my_script
  hago script get my_script -o json
  hago script get my_script --yaml > script.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		config, err := getClient().ScriptGet(ctx, args[0])
		if err != nil {
			return err
		}

		asYAML, _ := cmd.Flags().GetBool("yaml")
		if asYAML {
			return yaml.NewEncoder(os.Stdout).Encode(config)
		}

		return printResult(config)
	},
}

var scriptSaveCmd = &cobra.Command{
	Use:   "save <id>",
	Short: "Save script configuration",
	Long: `Save (create or update) a script configuration.

WARNING: This uses an undocumented REST API endpoint that may change without notice.

The configuration can be provided via file (-f) or stdin. It must be JSON or YAML
containing the full script configuration including id, alias, and sequence.

Examples:
  hago script save my_script -f script.yaml
  hago script save my_script -f script.json
  cat script.yaml | hago script save my_script`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		file, _ := cmd.Flags().GetString("file")

		// Read config from file or stdin
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
		var config hago.ScriptConfig
		if err := json.Unmarshal(data, &config); err != nil {
			// Try YAML
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("parse config (tried JSON and YAML): %w", err)
			}
		}

		// Ensure ID matches argument
		config.ID = args[0]

		if err := getClient().ScriptSave(ctx, &config); err != nil {
			return err
		}

		printSuccess("Script configuration '%s' saved", args[0])
		return nil
	},
}

var scriptDeleteConfigCmd = &cobra.Command{
	Use:   "delete-config <id>",
	Short: "Delete script configuration",
	Long: `Delete a script configuration by ID.

WARNING: This uses an undocumented REST API endpoint that may change without notice.
This removes the script configuration entirely.

Examples:
  hago script delete-config my_script`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().ScriptDeleteConfig(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Script configuration '%s' deleted", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)

	scriptCmd.AddCommand(scriptRunCmd)
	scriptCmd.AddCommand(scriptTurnOnCmd)
	scriptCmd.AddCommand(scriptTurnOffCmd)
	scriptCmd.AddCommand(scriptToggleCmd)
	scriptCmd.AddCommand(scriptReloadCmd)
	scriptCmd.AddCommand(scriptListCmd)
	scriptCmd.AddCommand(scriptGetCmd)
	scriptCmd.AddCommand(scriptSaveCmd)
	scriptCmd.AddCommand(scriptDeleteConfigCmd)

	// Run flags
	scriptRunCmd.Flags().String("vars", "", "Variables as JSON (e.g., '{\"key\":\"value\"}')")

	// Turn-on flags
	scriptTurnOnCmd.Flags().String("vars", "", "Variables as JSON (e.g., '{\"key\":\"value\"}')")

	// Get flags
	scriptGetCmd.Flags().Bool("yaml", false, "Output as YAML")

	// Save flags
	scriptSaveCmd.Flags().StringP("file", "f", "", "File containing script configuration (JSON or YAML)")
}
