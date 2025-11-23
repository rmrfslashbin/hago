package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

// outputFormat controls how results are displayed.
var outputFormat string

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, pretty)")
}

// printResult outputs the result in the configured format.
func printResult(v any) error {
	switch outputFormat {
	case "pretty":
		return printPrettyJSON(v)
	default:
		return printJSON(v)
	}
}

// printJSON outputs compact JSON.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(v)
}

// printPrettyJSON outputs indented JSON.
func printPrettyJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printSuccess outputs a success message.
func printSuccess(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
