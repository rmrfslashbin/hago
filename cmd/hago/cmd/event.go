package cmd

import (
	"fmt"

	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Manage events",
	Long:  `List event types or fire events.`,
}

var eventListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List event types",
	Long:    `List all event types with their listener counts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		events, err := getClient().Events(ctx)
		if err != nil {
			return err
		}
		return printResult(events)
	},
}

var eventFireCmd = &cobra.Command{
	Use:   "fire <event_type>",
	Short: "Fire an event",
	Long: `Fire a custom event with optional data.

Examples:
  hago event fire my_custom_event
  hago event fire my_custom_event --data '{"key": "value"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		eventType := args[0]

		var data hago.EventData
		dataJSON, _ := cmd.Flags().GetString("data")
		if dataJSON != "" {
			parsed, err := parseJSON(dataJSON)
			if err != nil {
				return fmt.Errorf("invalid data JSON: %w", err)
			}
			data = parsed
		}

		if err := getClient().FireEvent(ctx, eventType, data); err != nil {
			return err
		}
		printSuccess("Event '%s' fired successfully", eventType)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventFireCmd)

	eventFireCmd.Flags().StringP("data", "d", "", "Event data as JSON")
}
