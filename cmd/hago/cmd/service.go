package cmd

import (
	"fmt"

	"github.com/rmrfslashbin/hago"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	Long:  `List available services or call a service.`,
}

var serviceListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available services",
	Long:    `List all services available in Home Assistant, grouped by domain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		services, err := getClient().Services(ctx)
		if err != nil {
			return err
		}
		return printResult(services)
	},
}

var serviceCallCmd = &cobra.Command{
	Use:   "call <domain> <service> [entity_id]",
	Short: "Call a service",
	Long: `Call a Home Assistant service.

Examples:
  hago service call light turn_on light.living_room
  hago service call light turn_on light.living_room --data '{"brightness": 255}'
  hago service call homeassistant restart`,
	Args: cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		domain := args[0]
		service := args[1]

		req := &hago.ServiceCallRequest{}

		if len(args) > 2 {
			req.EntityID = args[2]
		}

		dataJSON, _ := cmd.Flags().GetString("data")
		if dataJSON != "" {
			data, err := parseJSON(dataJSON)
			if err != nil {
				return fmt.Errorf("invalid data JSON: %w", err)
			}
			req.Data = data
		}

		states, err := getClient().CallService(ctx, domain, service, req)
		if err != nil {
			return err
		}
		return printResult(states)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceCallCmd)

	serviceCallCmd.Flags().StringP("data", "d", "", "Service data as JSON")
}
