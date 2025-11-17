package toon

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Options for TOON encoding/decoding
type Options struct {
	Indent    int
	Delimiter string
	UseTab    bool
}

// DefaultOptions returns default TOON options
func DefaultOptions() Options {
	return Options{
		Indent:    2,
		Delimiter: ",",
		UseTab:    false,
	}
}

// Value represents a TOON value
type Value interface{}

// Encode converts a Go value to TOON format
func Encode(v interface{}, opts Options) (string, error) {
	return encode(v, opts, 0)
}

func encode(v interface{}, opts Options, depth int) (string, error) {
	if v == nil {
		return "null", nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		return encodeObject(val, opts, depth)
	case []interface{}:
		return encodeArray(val, opts, depth)
	case bool:
		return fmt.Sprintf("%t", val), nil
	case float64, int, int64:
		return fmt.Sprintf("%v", val), nil
	case string:
		return encodeString(val, opts.Delimiter), nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

func encodeObject(obj map[string]interface{}, opts Options, depth int) (string, error) {
	if len(obj) == 0 {
		return "", nil
	}

	var lines []string
	indent := makeIndent(depth, opts)

	// Sort keys for deterministic output
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := obj[key]
		if value == nil {
			lines = append(lines, fmt.Sprintf("%s%s: null", indent, key))
			continue
		}

		switch v := value.(type) {
		case map[string]interface{}:
			nested, err := encodeObject(v, opts, depth+1)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("%s%s:", indent, key))
			lines = append(lines, nested)

		case []interface{}:
			encoded, err := encodeArrayValue(v, opts, depth+1)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("%s%s%s", indent, key, encoded))

		default:
			encoded, err := encode(value, opts, depth)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("%s%s: %s", indent, key, encoded))
		}
	}

	return strings.Join(lines, "\n"), nil
}

func encodeArray(arr []interface{}, opts Options, depth int) (string, error) {
	if len(arr) == 0 {
		return "[0]:", nil
	}

	// Check if array is uniform (all objects with same keys)
	if isUniformArray(arr) {
		return encodeTabularArray(arr, opts, depth)
	}

	// Check if all primitives
	if isAllPrimitives(arr) {
		return encodePrimitiveArray(arr, opts)
	}

	// Mixed array - use list format
	return encodeMixedArray(arr, opts, depth)
}

// encodeArrayValue is an alias for encodeArray for backward compatibility
func encodeArrayValue(arr []interface{}, opts Options, depth int) (string, error) {
	return encodeArray(arr, opts, depth)
}

func encodePrimitiveArray(arr []interface{}, opts Options) (string, error) {
	var values []string
	for _, item := range arr {
		switch v := item.(type) {
		case string:
			values = append(values, encodeString(v, opts.Delimiter))
		case bool:
			values = append(values, fmt.Sprintf("%t", v))
		default:
			values = append(values, fmt.Sprintf("%v", v))
		}
	}
	return fmt.Sprintf("[%d]: %s", len(arr), strings.Join(values, opts.Delimiter)), nil
}

func encodeTabularArray(arr []interface{}, opts Options, depth int) (string, error) {
	if len(arr) == 0 {
		return "[0]:", nil
	}

	// Get fields from first object
	first, ok := arr[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected object in uniform array")
	}

	// Sort fields for deterministic output
	var fields []string
	for key := range first {
		fields = append(fields, key)
	}
	sort.Strings(fields)

	// Build header
	header := fmt.Sprintf("[%d]{%s}:", len(arr), strings.Join(fields, opts.Delimiter))

	// Build rows
	indent := makeIndent(depth, opts)
	var rows []string
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var values []string
		for _, field := range fields {
			val := obj[field]
			switch v := val.(type) {
			case string:
				values = append(values, encodeString(v, opts.Delimiter))
			case nil:
				values = append(values, "null")
			default:
				values = append(values, fmt.Sprintf("%v", v))
			}
		}
		rows = append(rows, fmt.Sprintf("%s%s", indent, strings.Join(values, opts.Delimiter)))
	}

	return header + "\n" + strings.Join(rows, "\n"), nil
}

func encodeMixedArray(arr []interface{}, opts Options, depth int) (string, error) {
	indent := makeIndent(depth, opts)
	var lines []string
	lines = append(lines, fmt.Sprintf("[%d]:", len(arr)))

	for _, item := range arr {
		encoded, err := encode(item, opts, depth)
		if err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("%s- %s", indent, encoded))
	}

	return strings.Join(lines, "\n"), nil
}

func encodeString(s, delimiter string) string {
	// Quote if needed
	needsQuote := strings.HasPrefix(s, " ") ||
		strings.HasSuffix(s, " ") ||
		strings.Contains(s, delimiter) ||
		strings.Contains(s, ":") ||
		strings.Contains(s, "\n") ||
		s == "true" || s == "false" || s == "null" ||
		strings.HasPrefix(s, "-") ||
		looksLikeNumber(s)

	if needsQuote {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(s, "\"", "\\\""))
	}
	return s
}

func looksLikeNumber(s string) bool {
	// Check if string is a valid number and nothing else
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isUniformArray(arr []interface{}) bool {
	if len(arr) == 0 {
		return false
	}

	first, ok := arr[0].(map[string]interface{})
	if !ok {
		return false
	}

	// Get first object's keys
	var firstKeys []string
	for key := range first {
		firstKeys = append(firstKeys, key)
	}

	// Check all other objects have same keys
	for i := 1; i < len(arr); i++ {
		obj, ok := arr[i].(map[string]interface{})
		if !ok {
			return false
		}

		if len(obj) != len(firstKeys) {
			return false
		}

		for _, key := range firstKeys {
			if _, exists := obj[key]; !exists {
				return false
			}
		}
	}

	return true
}

func isAllPrimitives(arr []interface{}) bool {
	for _, item := range arr {
		switch item.(type) {
		case map[string]interface{}, []interface{}:
			return false
		}
	}
	return true
}

func makeIndent(depth int, opts Options) string {
	if depth == 0 {
		return ""
	}
	if opts.UseTab {
		return strings.Repeat("\t", depth)
	}
	return strings.Repeat(" ", depth*opts.Indent)
}
