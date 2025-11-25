package cmd

import (
	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
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

func init() {
	rootCmd.AddCommand(automationCmd)

	automationCmd.AddCommand(automationTriggerCmd)
	automationCmd.AddCommand(automationTurnOnCmd)
	automationCmd.AddCommand(automationTurnOffCmd)
	automationCmd.AddCommand(automationToggleCmd)
	automationCmd.AddCommand(automationReloadCmd)

	// Trigger flags
	automationTriggerCmd.Flags().Bool("skip-condition", false, "Skip automation conditions when triggering")
}
