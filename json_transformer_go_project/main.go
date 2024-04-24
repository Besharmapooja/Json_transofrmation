package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

func main() {
	data, err := ioutil.ReadFile("input.json")
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	var inputData interface{}
	err = json.Unmarshal(data, &inputData)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	transformedData, err := TransformJSON(inputData)
	if err != nil {
		log.Fatalf("Error transforming JSON: %v", err)
	}

	output, err := json.MarshalIndent(transformedData, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	fmt.Println(string(output))
}

func TransformJSON(input interface{}) (interface{}, error) {
	switch v := input.(type) {
	case map[string]interface{}:
		return transformMap(v)
	case []interface{}:
		return transformList(v) // Handle lists at any nesting level
	default:
		return nil, fmt.Errorf("input must be a JSON object or array")
	}
}

func transformMap(input map[string]interface{}) (map[string]interface{}, error) {
	transformed := make(map[string]interface{})
	for key, value := range input {
		sanitizedKey := strings.TrimSpace(key)
		if sanitizedKey == "" {
			continue
		}
		if v, ok := value.(map[string]interface{}); ok {
			transformedValue, err := handleDataTypes(sanitizedKey, v)
			if err == nil && transformedValue != nil {
				transformed[sanitizedKey] = transformedValue
			}
		}
	}
	return transformed, nil
}

func handleDataTypes(key string, value map[string]interface{}) (interface{}, error) {
	for typeKey, typeValue := range value {
		switch strings.TrimSpace(typeKey) {
		case "S":
			if str, ok := typeValue.(string); ok {
				return transformString(str)
			}
		case "N":
			if str, ok := typeValue.(string); ok {
				return transformNumber(str)
			}
		case "BOOL":
			if str, ok := typeValue.(string); ok {
				transformedBool, err := transformBool(str)
				if err != nil || transformedBool == nil { // Check for nil or error
					return nil, nil // Explicitly return nil to avoid adding to the output
				}
				return transformedBool, nil
			}
		case "NULL":
			if str, ok := typeValue.(string); ok {
				transformed, err := transformNull(str)
				if transformed == nil && err == nil { // Explicitly check for nil as a valid transformed value
					return nil, nil // Return nil explicitly to ensure it gets included
				}
				return transformed, err
			}
		case "L":
			if list, ok := typeValue.([]interface{}); ok {
				return transformList(list)
			}
		case "M":
			if m, ok := typeValue.(map[string]interface{}); ok {
				return transformMap(m)
			}
		}
	}
	return nil, fmt.Errorf("invalid type or unsupported key: %s", key)
}

func transformString(input string) (interface{}, error) {
	sanitized := strings.TrimSpace(input)
	if sanitized == "" || sanitized == "noop" {
		return nil, nil // Filter out empty strings and "noop"
	}
	if t, err := time.Parse(time.RFC3339, sanitized); err == nil {
		return t.Unix(), nil
	}
	return sanitized, nil
}

func transformNumber(input string) (interface{}, error) {
	sanitized := strings.TrimLeft(strings.TrimSpace(input), "0")
	if sanitized == "" || !isValidNumber(sanitized) { // Implement isValidNumber to check numeric validity
		return nil, nil
	}
	if i, err := strconv.Atoi(sanitized); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(sanitized, 64); err == nil {
		return f, nil
	}
	return nil, nil
}

func isValidNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func transformBool(input string) (interface{}, error) {
	sanitized := strings.ToLower(strings.TrimSpace(input))
	switch sanitized {
	case "1", "t", "true":
		return true, nil
	case "0", "f", "false":
		return false, nil
	default:
		return nil, nil // Return nil for any non-valid boolean strings
	}
}

func transformNull(input string) (interface{}, error) {
	sanitized := strings.ToLower(strings.TrimSpace(input))
	switch sanitized {
	case "true", "t", "1":
		return nil, nil
	}
	return nil, nil // Always return nil to handle other cases as null
}

func transformList(input []interface{}) ([]interface{}, error) {
	var list []interface{}
	for _, item := range input {
		if itemMap, ok := item.(map[string]interface{}); ok {
			for key, val := range itemMap {
				switch key {
				case "N": // Handling numeric values
					if num, err := transformNumber(val.(string)); err == nil && num != nil {
						list = append(list, num)
					}
				case "BOOL": // Handling boolean values
					if boo, err := transformBool(val.(string)); err == nil && boo != nil {
						list = append(list, boo)
					}
				case "S": // Handling String values
					if s, err := transformString(val.(string)); err == nil && s != nil {
						list = append(list, s)
					}
				}
			}
		}
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("list is empty or contains no valid items")
	}
	return list, nil
}
