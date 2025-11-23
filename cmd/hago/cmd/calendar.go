package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Manage calendars",
	Long:  `List calendars or get calendar events.`,
}

var calendarListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List calendars",
	Long:    `List all calendar entities in Home Assistant.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		calendars, err := getClient().Calendars(ctx)
		if err != nil {
			return err
		}
		return printResult(calendars)
	},
}

var calendarEventsCmd = &cobra.Command{
	Use:   "events <entity_id>",
	Short: "Get calendar events",
	Long: `Get events from a calendar entity.

Examples:
  hago calendar events calendar.personal
  hago calendar events calendar.work --days 14`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entityID := args[0]

		days, _ := cmd.Flags().GetInt("days")
		start := time.Now()
		end := start.AddDate(0, 0, days)

		events, err := getClient().CalendarEvents(ctx, entityID, start, end)
		if err != nil {
			return err
		}
		return printResult(events)
	},
}

func init() {
	rootCmd.AddCommand(calendarCmd)
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarEventsCmd)

	calendarEventsCmd.Flags().IntP("days", "d", 7, "Number of days to fetch events for")
}
