package cmd

import (
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Query Home Assistant registries",
	Long: `Query Home Assistant registry APIs for metadata about entities, devices, areas, labels, and floors.

Registries store organizational metadata that isn't available through the state API,
such as area assignments, device information, and user-defined labels.`,
}

var entityRegistryCmd = &cobra.Command{
	Use:     "entities",
	Aliases: []string{"entity", "ent"},
	Short:   "List entity registry entries",
	Long: `List all entities in the registry with metadata including:
- Area assignments
- Device associations
- Labels and categories
- Custom icons and names
- Disabled/hidden status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entries, err := getClient().EntityRegistry(ctx)
		if err != nil {
			return err
		}
		return printResult(entries)
	},
}

var deviceRegistryCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device", "dev"},
	Short:   "List device registry entries",
	Long: `List all devices in the registry with metadata including:
- Manufacturer, model, and hardware details
- Firmware/software versions
- Area assignments
- Device connections and identifiers
- Labels`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entries, err := getClient().DeviceRegistry(ctx)
		if err != nil {
			return err
		}
		return printResult(entries)
	},
}

var areaRegistryCmd = &cobra.Command{
	Use:     "areas",
	Aliases: []string{"area"},
	Short:   "List area registry entries",
	Long: `List all areas (rooms/locations) in the registry with metadata including:
- Floor assignments
- Icons and pictures
- Aliases
- Labels`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entries, err := getClient().AreaRegistry(ctx)
		if err != nil {
			return err
		}
		return printResult(entries)
	},
}

var labelRegistryCmd = &cobra.Command{
	Use:     "labels",
	Aliases: []string{"label"},
	Short:   "List label registry entries",
	Long: `List all labels (user-defined organizational tags) in the registry with metadata including:
- Icons and colors
- Descriptions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entries, err := getClient().LabelRegistry(ctx)
		if err != nil {
			return err
		}
		return printResult(entries)
	},
}

var floorRegistryCmd = &cobra.Command{
	Use:     "floors",
	Aliases: []string{"floor"},
	Short:   "List floor registry entries",
	Long: `List all floors (building levels) in the registry with metadata including:
- Level numbers
- Icons
- Aliases`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		entries, err := getClient().FloorRegistry(ctx)
		if err != nil {
			return err
		}
		return printResult(entries)
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)

	registryCmd.AddCommand(entityRegistryCmd)
	registryCmd.AddCommand(deviceRegistryCmd)
	registryCmd.AddCommand(areaRegistryCmd)
	registryCmd.AddCommand(labelRegistryCmd)
	registryCmd.AddCommand(floorRegistryCmd)
}
