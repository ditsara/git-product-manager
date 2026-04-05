package ticket

import (
	"fmt"
	"strconv"
	"strings"
)

// FieldType represents the data type of a ticket field
type FieldType int

const (
	FieldTypeString FieldType = iota
	FieldTypeInt
	FieldTypeArray
	FieldTypeEnum
)

// FieldSpec defines the specification for a ticket field
type FieldSpec struct {
	Type          FieldType
	AllowedValues []string // For enum fields
}

// FieldRegistry maps field names to their specifications
var FieldRegistry = map[string]FieldSpec{
	// Array fields
	"labels":      {Type: FieldTypeArray},
	"milestones":  {Type: FieldTypeArray},
	"depends_on": {Type: FieldTypeArray},
	"blocks":     {Type: FieldTypeArray},
	"related":    {Type: FieldTypeArray},

	// Integer fields
	"points": {Type: FieldTypeInt},

	// String fields
	"id":         {Type: FieldTypeString},
	"title":      {Type: FieldTypeString},
	"assignee":   {Type: FieldTypeString},
	"parent":     {Type: FieldTypeString},
	"created_at": {Type: FieldTypeString},
	"updated_at": {Type: FieldTypeString},

	// Enum fields (validation handled separately for status/type/priority)
	"status":   {Type: FieldTypeEnum}, // Validated against workflow.yaml
	"type":     {Type: FieldTypeEnum, AllowedValues: []string{"epic", "story", "task", "bug"}},
	"priority": {Type: FieldTypeEnum, AllowedValues: []string{"low", "medium", "high", "critical"}},
}

// ParseFieldValue parses a string value according to the field's type
func ParseFieldValue(fieldName, value string) (interface{}, error) {
	spec, ok := FieldRegistry[fieldName]
	if !ok {
		// Unknown field - treat as string
		return value, nil
	}

	switch spec.Type {
	case FieldTypeArray:
		return parseArrayValue(value), nil

	case FieldTypeInt:
		if value == "" {
			return 0, nil
		}
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("field '%s' must be an integer, got '%s'", fieldName, value)
		}
		return intVal, nil

	case FieldTypeEnum:
		if len(spec.AllowedValues) > 0 {
			// Validate against allowed values
			for _, allowed := range spec.AllowedValues {
				if value == allowed {
					return value, nil
				}
			}
			return nil, fmt.Errorf("invalid value '%s' for field '%s'. Allowed values: %v",
				value, fieldName, spec.AllowedValues)
		}
		// status is validated against workflow.yaml elsewhere
		return value, nil

	case FieldTypeString:
		return value, nil

	default:
		return value, nil
	}
}

// AppendDomain appends a domain suffix to a username if:
//   - domain is non-empty
//   - username does not already contain "@"
func AppendDomain(username, domain string) string {
	if domain == "" || strings.Contains(username, "@") {
		return username
	}
	return username + "@" + domain
}

// parseArrayValue parses a comma-separated string into a string slice
// Handles edge cases:
// - Trims whitespace from each element
// - Filters out empty elements
// - Empty string returns empty array
func parseArrayValue(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
