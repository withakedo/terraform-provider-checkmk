package labels

import (
	"encoding/json"
	"fmt"
)

// parseLabelsFromValueRaw parses the value_raw string back to a labels map.
// The API returns value_raw in Python dict format: "{'key': 'value'}"
func parseLabelsFromValueRaw(valueRaw string) (map[string]string, error) {
	// Try JSON first
	var labelsMap map[string]string
	if err := json.Unmarshal([]byte(valueRaw), &labelsMap); err == nil {
		return labelsMap, nil
	}

	// API returns Python dict format, need to convert single quotes to double quotes
	// This is a simple conversion that handles basic cases
	converted := convertPythonDictToJSON(valueRaw)
	if err := json.Unmarshal([]byte(converted), &labelsMap); err != nil {
		return nil, fmt.Errorf("failed to parse value_raw '%s': %w", valueRaw, err)
	}
	return labelsMap, nil
}

// convertPythonDictToJSON converts Python dict format to JSON
func convertPythonDictToJSON(pythonDict string) string {
	// Simple conversion: replace single quotes with double quotes
	// This handles basic cases like "{'key': 'value'}" -> '{"key": "value"}'
	result := ""
	inString := false
	for i := 0; i < len(pythonDict); i++ {
		c := pythonDict[i]
		if c == '\'' {
			if !inString {
				result += "\""
				inString = true
			} else {
				result += "\""
				inString = false
			}
		} else {
			result += string(c)
		}
	}
	return result
}
