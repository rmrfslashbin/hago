package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check API status",
	Long:  `Check if the Home Assistant API is running and responding.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		status, err := getClient().Status(ctx)
		if err != nil {
			return err
		}
		return printResult(status)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
