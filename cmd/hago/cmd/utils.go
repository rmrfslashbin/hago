package cmd

import (
	"encoding/json"
	"fmt"
)

// parseJSON parses a JSON string into a map.
func parseJSON(s string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return result, nil
}
