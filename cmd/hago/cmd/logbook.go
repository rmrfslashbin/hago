package cmd

import (
	"fmt"
	"time"

	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
)

var logbookCmd = &cobra.Command{
	Use:   "logbook",
	Short: "Get logbook entries",
	Long: `Get logbook entries for Home Assistant.

Examples:
  hago logbook
  hago logbook --duration 12h
  hago logbook --entity light.living_room`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		durationStr, _ := cmd.Flags().GetString("duration")
		duration, err := parseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}

		entity, _ := cmd.Flags().GetString("entity")

		startTime := time.Now().Add(-duration)
		var opts *hago.LogbookOptions
		if entity != "" {
			opts = &hago.LogbookOptions{
				Entity: entity,
			}
		}

		entries, err := getClient().Logbook(ctx, startTime, opts)
		if err != nil {
			return err
		}
		return printResult(entries)
	},
}

func init() {
	rootCmd.AddCommand(logbookCmd)

	logbookCmd.Flags().StringP("duration", "d", "24h", "Logbook duration (e.g., 24h, 7d)")
	logbookCmd.Flags().StringP("entity", "e", "", "Filter by entity ID")
}
