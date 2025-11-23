package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template <template>",
	Short: "Render a Jinja2 template",
	Long: `Render a Jinja2 template using Home Assistant's template engine.

Examples:
  hago template "{{ states('light.living_room') }}"
  hago template "{{ state_attr('light.living_room', 'brightness') }}"
  hago template "{{ now().strftime('%Y-%m-%d') }}"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		result, err := getClient().RenderTemplate(ctx, args[0])
		if err != nil {
			return err
		}
		fmt.Print(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
