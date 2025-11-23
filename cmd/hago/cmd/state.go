package cmd

import (
	"fmt"

	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
)

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Manage entity states",
	Long:  `Get, set, or delete entity states in Home Assistant.`,
}

var stateListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all entity states",
	Long:    `List all entity states currently tracked by Home Assistant.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		states, err := getClient().States(ctx)
		if err != nil {
			return err
		}
		return printResult(states)
	},
}

var stateGetCmd = &cobra.Command{
	Use:   "get <entity_id>",
	Short: "Get state of an entity",
	Long:  `Get the current state and attributes of a specific entity.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		state, err := getClient().State(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(state)
	},
}

var stateSetCmd = &cobra.Command{
	Use:   "set <entity_id> <state>",
	Short: "Set state of an entity",
	Long: `Create or update the state of an entity.

Examples:
  hago state set sensor.test 42
  hago state set sensor.test "on" --attr '{"unit": "celsius"}'`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		attrJSON, _ := cmd.Flags().GetString("attr")

		update := &hago.StateUpdate{
			State: args[1],
		}

		if attrJSON != "" {
			attrs, err := parseJSON(attrJSON)
			if err != nil {
				return fmt.Errorf("invalid attributes JSON: %w", err)
			}
			update.Attributes = attrs
		}

		state, err := getClient().SetState(ctx, args[0], update)
		if err != nil {
			return err
		}
		return printResult(state)
	},
}

var stateDeleteCmd = &cobra.Command{
	Use:     "delete <entity_id>",
	Aliases: []string{"rm"},
	Short:   "Delete an entity",
	Long:    `Remove an entity from Home Assistant.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := getClient().DeleteState(ctx, args[0]); err != nil {
			return err
		}
		printSuccess("Entity %s deleted", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(stateListCmd)
	stateCmd.AddCommand(stateGetCmd)
	stateCmd.AddCommand(stateSetCmd)
	stateCmd.AddCommand(stateDeleteCmd)

	stateSetCmd.Flags().String("attr", "", "Attributes as JSON")
}
