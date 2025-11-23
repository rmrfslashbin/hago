package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Get Home Assistant configuration",
	Long:  `Retrieve the current Home Assistant configuration including version, location, and units.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		config, err := getClient().Config(ctx)
		if err != nil {
			return err
		}
		return printResult(config)
	},
}

var checkConfigCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate Home Assistant configuration",
	Long:  `Check if the Home Assistant configuration.yaml is valid.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		result, err := getClient().CheckConfig(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(checkConfigCmd)
}
