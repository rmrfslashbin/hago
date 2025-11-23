package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var errorlogCmd = &cobra.Command{
	Use:   "errorlog",
	Short: "Get error log",
	Long:  `Retrieve the Home Assistant error log.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		log, err := getClient().ErrorLog(ctx)
		if err != nil {
			return err
		}
		fmt.Print(log)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(errorlogCmd)
}
