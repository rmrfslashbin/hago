package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var componentsCmd = &cobra.Command{
	Use:   "components",
	Short: "List loaded components",
	Long:  `List all components currently loaded in Home Assistant.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		components, err := getClient().Components(ctx)
		if err != nil {
			return err
		}

		if outputFormat == "json" || outputFormat == "pretty" {
			return printResult(components)
		}

		for _, c := range components {
			fmt.Println(c)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(componentsCmd)
}
