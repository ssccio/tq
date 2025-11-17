package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Engine executes queries on data
type Engine struct{}

// New creates a new query engine
func New() *Engine {
	return &Engine{}
}

// Execute runs a query on the given data
func (e *Engine) Execute(query string, data interface{}) (interface{}, error) {
	query = strings.TrimSpace(query)

	// Handle identity
	if query == "." {
		return data, nil
	}

	// Parse and execute query
	return e.executeQuery(query, data)
}

func (e *Engine) executeQuery(query string, data interface{}) (interface{}, error) {
	// Handle pipe operations
	if strings.Contains(query, "|") {
		return e.executePipe(query, data)
	}

	// Handle array operations
	if strings.Contains(query, "[]") {
		return e.executeArrayIteration(query, data)
	}

	// Handle field access with dots
	if strings.HasPrefix(query, ".") && !strings.Contains(query, "(") {
		return e.executeFieldAccess(query, data)
	}

	// Handle select operations
	if strings.HasPrefix(query, "select(") {
		return e.executeSelect(query, data)
	}

	// Handle array construction
	if strings.HasPrefix(query, "[") && strings.HasSuffix(query, "]") {
		return e.executeArrayConstruction(query, data)
	}

	// Handle object construction
	if strings.HasPrefix(query, "{") && strings.HasSuffix(query, "}") {
		return e.executeObjectConstruction(query, data)
	}

	return nil, fmt.Errorf("unsupported query: %s", query)
}

func (e *Engine) executePipe(query string, data interface{}) (interface{}, error) {
	// Split by pipe, handling nested structures
	parts := splitPipe(query)

	result := data
	for _, part := range parts {
		var err error
		result, err = e.executeQuery(strings.TrimSpace(part), result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (e *Engine) executeFieldAccess(query string, data interface{}) (interface{}, error) {
	// Remove leading dot
	path := strings.TrimPrefix(query, ".")

	// Handle array index access like .[0] or .items[0]
	if strings.Contains(path, "[") {
		return e.executeArrayIndex(path, data)
	}

	// Split by dots for nested access
	parts := strings.Split(path, ".")

	current := data
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle map
		if m, ok := current.(map[string]interface{}); ok {
			var exists bool
			current, exists = m[part]
			if !exists {
				return nil, nil // Field doesn't exist
			}
		} else {
			return nil, fmt.Errorf("cannot access field '%s' on non-object", part)
		}
	}

	return current, nil
}

func (e *Engine) executeArrayIndex(path string, data interface{}) (interface{}, error) {
	// Parse path like "items[0]" or "[1]" or "items[0].name"
	// First, find the bracket
	bracketStart := strings.Index(path, "[")
	if bracketStart == -1 {
		return nil, fmt.Errorf("no array index found in path: %s", path)
	}

	bracketEnd := strings.Index(path, "]")
	if bracketEnd == -1 {
		return nil, fmt.Errorf("unclosed bracket in path: %s", path)
	}

	// Extract parts
	fieldPart := ""
	if bracketStart > 0 {
		fieldPart = path[:bracketStart]
	}
	indexStr := path[bracketStart+1 : bracketEnd]
	remainingPath := ""
	if bracketEnd+1 < len(path) {
		remainingPath = path[bracketEnd+1:]
		// Remove leading dot if present
		remainingPath = strings.TrimPrefix(remainingPath, ".")
	}

	// Get the array
	var arr []interface{}
	if fieldPart != "" {
		// Access field first
		result, err := e.executeFieldAccess("."+fieldPart, data)
		if err != nil {
			return nil, err
		}
		var ok bool
		arr, ok = result.([]interface{})
		if !ok {
			return nil, fmt.Errorf("field '%s' is not an array", fieldPart)
		}
	} else {
		var ok bool
		arr, ok = data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("data is not an array")
		}
	}

	// Parse index
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array index '%s': %w", indexStr, err)
	}

	// Handle negative indices
	if index < 0 {
		index = len(arr) + index
	}

	if index < 0 || index >= len(arr) {
		return nil, fmt.Errorf("array index out of bounds: %d (array length: %d)", index, len(arr))
	}

	result := arr[index]

	// If there's a remaining path, continue accessing
	if remainingPath != "" {
		return e.executeFieldAccess("."+remainingPath, result)
	}

	return result, nil
}

func (e *Engine) executeArrayIteration(query string, data interface{}) (interface{}, error) {
	// Parse query like ".items[]" or ".[]"
	query = strings.TrimSpace(query)

	// Get the array
	var arr []interface{}
	if query == ".[]" {
		var ok bool
		arr, ok = data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("data is not an array")
		}
	} else {
		// Extract field path
		fieldPath := strings.TrimSuffix(query, "[]")
		result, err := e.executeFieldAccess(fieldPath, data)
		if err != nil {
			return nil, err
		}
		var ok bool
		arr, ok = result.([]interface{})
		if !ok {
			return nil, fmt.Errorf("field is not an array")
		}
	}

	// Return array elements (will be handled by caller for iteration)
	return arr, nil
}

func (e *Engine) executeSelect(query string, data interface{}) (interface{}, error) {
	// Parse select(condition)
	if !strings.HasPrefix(query, "select(") || !strings.HasSuffix(query, ")") {
		return nil, fmt.Errorf("invalid select syntax")
	}

	condition := query[7 : len(query)-1]

	// Evaluate condition
	result, err := e.evaluateCondition(condition, data)
	if err != nil {
		return nil, err
	}

	if result {
		return data, nil
	}

	return nil, nil
}

func (e *Engine) evaluateCondition(condition string, data interface{}) (bool, error) {
	// Handle simple comparisons like ".age > 25"
	operators := []string{">=", "<=", "==", "!=", ">", "<"}

	for _, op := range operators {
		if strings.Contains(condition, op) {
			parts := strings.SplitN(condition, op, 2)
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Evaluate left side
			leftVal, err := e.executeQuery(left, data)
			if err != nil {
				return false, err
			}

			// Parse right side
			rightVal, err := parseValue(right)
			if err != nil {
				return false, err
			}

			return compareValues(leftVal, rightVal, op)
		}
	}

	return false, fmt.Errorf("unsupported condition: %s", condition)
}

func (e *Engine) executeArrayConstruction(query string, data interface{}) (interface{}, error) {
	// Remove brackets
	inner := strings.TrimPrefix(strings.TrimSuffix(query, "]"), "[")

	if inner == "" {
		return []interface{}{}, nil
	}

	// Execute inner query
	result, err := e.executeQuery(inner, data)
	if err != nil {
		return nil, err
	}

	// Wrap in array if not already
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}

	return []interface{}{result}, nil
}

func (e *Engine) executeObjectConstruction(query string, data interface{}) (interface{}, error) {
	// Simple object construction: {key: value, ...}
	inner := strings.TrimPrefix(strings.TrimSuffix(query, "}"), "{")

	obj := make(map[string]interface{})

	// Parse key-value pairs
	pairs := strings.Split(inner, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		valueExpr := strings.TrimSpace(parts[1])

		// Execute value expression
		value, err := e.executeQuery(valueExpr, data)
		if err != nil {
			return nil, err
		}

		obj[key] = value
	}

	return obj, nil
}

func splitPipe(query string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, ch := range query {
		switch ch {
		case '(', '[', '{':
			depth++
			current.WriteRune(ch)
		case ')', ']', '}':
			depth--
			current.WriteRune(ch)
		case '|':
			if depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	// Try boolean
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	// Try number
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num, nil
	}

	// String (remove quotes if present)
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return strings.Trim(s, `"`), nil
	}

	return s, nil
}

func compareValues(left, right interface{}, op string) (bool, error) {
	// Convert to comparable types
	leftNum, leftOk := toNumber(left)
	rightNum, rightOk := toNumber(right)

	if leftOk && rightOk {
		switch op {
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		case "==":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		}
	}

	// String comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch op {
	case "==":
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	}

	return false, fmt.Errorf("cannot compare values with operator %s", op)
}

func toNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}

	// Try reflection
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), true
	case reflect.Float32, reflect.Float64:
		return val.Float(), true
	}

	return 0, false
}
