package cmd

import (
	"fmt"
	"time"

	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history <entity_id>",
	Short: "Get state history",
	Long: `Get state history for an entity.

Examples:
  hago history light.living_room
  hago history light.living_room --duration 48h
  hago history sensor.temperature --duration 7d --minimal`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entityID := args[0]

		durationStr, _ := cmd.Flags().GetString("duration")
		duration, err := parseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}

		minimal, _ := cmd.Flags().GetBool("minimal")
		noAttr, _ := cmd.Flags().GetBool("no-attributes")

		startTime := time.Now().Add(-duration)
		opts := &hago.HistoryOptions{
			FilterEntityID:  entityID,
			MinimalResponse: minimal,
			NoAttributes:    noAttr,
		}

		history, err := getClient().History(ctx, startTime, opts)
		if err != nil {
			return err
		}
		return printResult(history)
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().StringP("duration", "d", "24h", "History duration (e.g., 24h, 7d)")
	historyCmd.Flags().Bool("minimal", false, "Return minimal response (state and last_changed only)")
	historyCmd.Flags().Bool("no-attributes", false, "Exclude attributes from response")
}

// parseDuration parses a duration string that can include "d" for days.
func parseDuration(s string) (time.Duration, error) {
	// Handle days suffix
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
