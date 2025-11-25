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

var automationCmd = &cobra.Command{
	Use:   "automation",
	Short: "Manage Home Assistant automations",
	Long: `Control automations using the automation service.

Provides commands to trigger, enable, disable, toggle, and reload automations.`,
}

var automationTriggerCmd = &cobra.Command{
	Use:   "trigger <entity_id>",
	Short: "Trigger an automation",
	Long: `Trigger an automation, optionally skipping conditions.

Examples:
  hago automation trigger automation.front_door_lock
  hago automation trigger automation.lights_on --skip-condition`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		skipCondition, _ := cmd.Flags().GetBool("skip-condition")

		var skipPtr *bool
		if cmd.Flags().Changed("skip-condition") {
			skipPtr = &skipCondition
		}

		req := &hago.AutomationTriggerRequest{
			EntityID:      args[0],
			SkipCondition: skipPtr,
		}

		if err := getClient().AutomationTrigger(ctx, req); err != nil {
			return err
		}

		printSuccess("Automation '%s' triggered", args[0])
		return nil
	},
}

var automationTurnOnCmd = &cobra.Command{
	Use:     "turn-on <entity_id>",
	Aliases: []string{"enable", "on"},
	Short:   "Enable an automation",
	Long: `Enable (turn on) an automation.

Examples:
  hago automation turn-on automation.front_door_lock
  hago automation enable automation.front_door_lock`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().AutomationTurnOn(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Automation '%s' enabled", args[0])
		return nil
	},
}

var automationTurnOffCmd = &cobra.Command{
	Use:     "turn-off <entity_id>",
	Aliases: []string{"disable", "off"},
	Short:   "Disable an automation",
	Long: `Disable (turn off) an automation.

Examples:
  hago automation turn-off automation.front_door_lock
  hago automation disable automation.front_door_lock`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().AutomationTurnOff(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Automation '%s' disabled", args[0])
		return nil
	},
}

var automationToggleCmd = &cobra.Command{
	Use:   "toggle <entity_id>",
	Short: "Toggle an automation's state",
	Long: `Toggle an automation between enabled and disabled.

Examples:
  hago automation toggle automation.front_door_lock`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().AutomationToggle(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Automation '%s' toggled", args[0])
		return nil
	},
}

var automationReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload all automations from YAML",
	Long: `Reload all automations from YAML configuration files.

This is useful after manually editing automation YAML files to apply changes
without restarting Home Assistant.

Examples:
  hago automation reload`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().AutomationReload(ctx); err != nil {
			return err
		}

		printSuccess("All automations reloaded from YAML configuration")
		return nil
	},
}

var automationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all automation configurations",
	Long: `List all automation configurations from Home Assistant.

WARNING: This uses an undocumented REST API endpoint that may change without notice.

Examples:
  hago automation list
  hago automation list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		configs, err := getClient().AutomationList(ctx)
		if err != nil {
			return err
		}

		return printResult(configs)
	},
}

var automationGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get automation configuration by ID",
	Long: `Get a specific automation configuration by its ID.

WARNING: This uses an undocumented REST API endpoint that may change without notice.

Examples:
  hago automation get my_automation
  hago automation get my_automation -o json
  hago automation get my_automation --yaml > automation.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		config, err := getClient().AutomationGet(ctx, args[0])
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

var automationSaveCmd = &cobra.Command{
	Use:   "save <id>",
	Short: "Save automation configuration",
	Long: `Save (create or update) an automation configuration.

WARNING: This uses an undocumented REST API endpoint that may change without notice.

The configuration can be provided via file (-f) or stdin. It must be JSON or YAML
containing the full automation configuration including id, alias, trigger, and action.

Examples:
  hago automation save my_automation -f automation.yaml
  hago automation save my_automation -f automation.json
  cat automation.yaml | hago automation save my_automation`,
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
		var config hago.AutomationConfig
		if err := json.Unmarshal(data, &config); err != nil {
			// Try YAML
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("parse config (tried JSON and YAML): %w", err)
			}
		}

		// Ensure ID matches argument
		config.ID = args[0]

		if err := getClient().AutomationSave(ctx, &config); err != nil {
			return err
		}

		printSuccess("Automation configuration '%s' saved", args[0])
		return nil
	},
}

var automationDeleteConfigCmd = &cobra.Command{
	Use:   "delete-config <id>",
	Short: "Delete automation configuration",
	Long: `Delete an automation configuration by ID.

WARNING: This uses an undocumented REST API endpoint that may change without notice.
This removes the automation configuration entirely.

Examples:
  hago automation delete-config my_automation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := getClient().AutomationDeleteConfig(ctx, args[0]); err != nil {
			return err
		}

		printSuccess("Automation configuration '%s' deleted", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(automationCmd)

	automationCmd.AddCommand(automationTriggerCmd)
	automationCmd.AddCommand(automationTurnOnCmd)
	automationCmd.AddCommand(automationTurnOffCmd)
	automationCmd.AddCommand(automationToggleCmd)
	automationCmd.AddCommand(automationReloadCmd)
	automationCmd.AddCommand(automationListCmd)
	automationCmd.AddCommand(automationGetCmd)
	automationCmd.AddCommand(automationSaveCmd)
	automationCmd.AddCommand(automationDeleteConfigCmd)

	// Trigger flags
	automationTriggerCmd.Flags().Bool("skip-condition", false, "Skip automation conditions when triggering")

	// Get flags
	automationGetCmd.Flags().Bool("yaml", false, "Output as YAML")

	// Save flags
	automationSaveCmd.Flags().StringP("file", "f", "", "File containing automation configuration (JSON or YAML)")
}
