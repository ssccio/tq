package query

import (
	"fmt"
	"math"
	"reflect"
	"sort"
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
	// Check for construction operators FIRST (highest precedence)
	// This prevents pipes inside brackets from being split incorrectly
	if strings.HasPrefix(query, "[") && strings.HasSuffix(query, "]") {
		return e.executeArrayConstruction(query, data)
	}

	// Handle if-then-else
	if strings.HasPrefix(query, "if ") {
		return e.executeIf(query, data)
	}

	// Handle alternative operator //
	if strings.Contains(query, "//") {
		return e.executeAlternative(query, data)
	}

	if strings.HasPrefix(query, "{") && strings.HasSuffix(query, "}") {
		return e.executeObjectConstruction(query, data)
	}

	// Handle pipe operations (but only if not inside brackets)
	if strings.Contains(query, "|") {
		return e.executePipe(query, data)
	}

	// Handle array operations
	if strings.Contains(query, "[]") {
		return e.executeArrayIteration(query, data)
	}

	// Handle select operations (before general functions)
	if strings.HasPrefix(query, "select(") {
		return e.executeSelect(query, data)
	}

	// Handle built-in functions
	if strings.Contains(query, "(") && !strings.HasPrefix(query, ".") {
		return e.executeFunction(query, data)
	}

	// Handle field access with dots
	if strings.HasPrefix(query, ".") && !strings.Contains(query, "(") {
		return e.executeFieldAccess(query, data)
	}

	// Handle literal values (true, false, numbers, strings)
	if val, err := parseValue(query); err == nil {
		return val, nil
	}

	return nil, fmt.Errorf("unsupported query: %s", query)
}

func (e *Engine) executePipe(query string, data interface{}) (interface{}, error) {
	// Split by pipe, handling nested structures
	parts := splitPipe(query)

	result := data
	for i, part := range parts {
		part = strings.TrimSpace(part)
		var err error

		// Check if previous result is an array from [] iteration
		if arr, ok := result.([]interface{}); ok && i > 0 && strings.Contains(parts[i-1], "[]") {
			// Apply this part to each element
			var results []interface{}
			for _, elem := range arr {
				elemResult, err := e.executeQuery(part, elem)
				if err != nil {
					return nil, err
				}
				// Only include non-nil results (for select filters)
				if elemResult != nil {
					results = append(results, elemResult)
				}
			}
			result = results
		} else {
			result, err = e.executeQuery(part, result)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (e *Engine) executeFieldAccess(query string, data interface{}) (interface{}, error) {
	// Remove leading dot
	path := strings.TrimPrefix(query, ".")

	// Handle array iteration with field access like .items[].name
	if strings.Contains(path, "[]") {
		// Split into before and after []
		parts := strings.SplitN(path, "[]", 2)
		beforeArray := parts[0]
		afterArray := ""
		if len(parts) > 1 {
			afterArray = strings.TrimPrefix(parts[1], ".")
		}

		// Get the array
		var arr []interface{}
		if beforeArray == "" {
			// Direct array iteration: .[].field
			var ok bool
			arr, ok = data.([]interface{})
			if !ok {
				return nil, fmt.Errorf("data is not an array")
			}
		} else {
			// Field then array: .items[].field
			result, err := e.executeFieldAccess("."+beforeArray, data)
			if err != nil {
				return nil, err
			}
			var ok bool
			arr, ok = result.([]interface{})
			if !ok {
				return nil, fmt.Errorf("field '%s' is not an array", beforeArray)
			}
		}

		// If there's a field after [], apply it to each element
		if afterArray != "" {
			var results []interface{}
			for _, elem := range arr {
				elemResult, err := e.executeFieldAccess("."+afterArray, elem)
				if err != nil {
					return nil, err
				}
				results = append(results, elemResult)
			}
			return results, nil
		}

		// Otherwise just return the array
		return arr, nil
	}

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

	// Execute inner query - this might produce multiple results
	result, err := e.executeQuery(inner, data)
	if err != nil {
		return nil, err
	}

	// If result is already an array from iteration (e.g., .items[]),
	// return it as-is (this is what jq does)
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}

	// Otherwise wrap single result in array
	return []interface{}{result}, nil
}

func (e *Engine) executeObjectConstruction(query string, data interface{}) (interface{}, error) {
	// Object construction: {key: valueExpr, ...} or {key} (shorthand for {key: .key})
	inner := strings.TrimPrefix(strings.TrimSuffix(query, "}"), "{")
	inner = strings.TrimSpace(inner)

	if inner == "" {
		return map[string]interface{}{}, nil
	}

	obj := make(map[string]interface{})

	// Parse key-value pairs (handle nested structures)
	pairs := splitByComma(inner)
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Check if it's key:value or just key (shorthand)
		if strings.Contains(pair, ":") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid object construction syntax: %s", pair)
			}

			key := strings.TrimSpace(parts[0])
			valueExpr := strings.TrimSpace(parts[1])

			// Execute value expression
			value, err := e.executeQuery(valueExpr, data)
			if err != nil {
				return nil, fmt.Errorf("object construction: evaluating '%s': %w", valueExpr, err)
			}

			obj[key] = value
		} else {
			// Shorthand: {name} is equivalent to {name: .name}
			key := pair
			value, err := e.executeFieldAccess("."+key, data)
			if err != nil {
				return nil, fmt.Errorf("object construction: accessing field '%s': %w", key, err)
			}
			obj[key] = value
		}
	}

	return obj, nil
}

// splitByComma splits a string by commas, respecting nested structures
func splitByComma(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '(', '[', '{':
			depth++
			current.WriteRune(ch)
		case ')', ']', '}':
			depth--
			current.WriteRune(ch)
		case ',':
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

	// Try null
	if s == "null" {
		return nil, nil
	}

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

// executeFunction handles built-in functions
func (e *Engine) executeFunction(query string, data interface{}) (interface{}, error) {
	// Parse function name and arguments
	parenIdx := strings.Index(query, "(")
	if parenIdx == -1 {
		return nil, fmt.Errorf("invalid function syntax: %s", query)
	}

	funcName := strings.TrimSpace(query[:parenIdx])
	argsStr := query[parenIdx+1:]
	if !strings.HasSuffix(argsStr, ")") {
		return nil, fmt.Errorf("unclosed function parenthesis: %s", query)
	}
	argsStr = strings.TrimSuffix(argsStr, ")")

	switch funcName {
	case "length":
		return e.funcLength(data)
	case "keys":
		return e.funcKeys(data)
	case "values":
		return e.funcValues(data)
	case "type":
		return e.funcType(data)
	case "sort":
		return e.funcSort(data)
	case "sort_by":
		return e.funcSortBy(argsStr, data)
	case "group_by":
		return e.funcGroupBy(argsStr, data)
	case "map":
		return e.funcMap(argsStr, data)
	case "reverse":
		return e.funcReverse(data)
	case "has":
		return e.funcHas(argsStr, data)
	case "in":
		return e.funcIn(argsStr, data)
	case "split":
		return e.funcSplit(argsStr, data)
	case "join":
		return e.funcJoin(argsStr, data)
	case "startswith":
		return e.funcStartsWith(argsStr, data)
	case "endswith":
		return e.funcEndsWith(argsStr, data)
	case "contains":
		return e.funcContains(argsStr, data)
	case "add":
		return e.funcAdd(data)
	case "min":
		return e.funcMin(data)
	case "max":
		return e.funcMax(data)
	case "floor":
		return e.funcFloor(data)
	case "ceil":
		return e.funcCeil(data)
	case "round":
		return e.funcRound(data)
	case "unique":
		return e.funcUnique(data)
	case "flatten":
		return e.funcFlatten(argsStr, data)
	case "range":
		return e.funcRange(argsStr, data)
	case "first":
		return e.funcFirst(argsStr, data)
	case "last":
		return e.funcLast(argsStr, data)
	case "tostring":
		return e.funcToString(data)
	case "tonumber":
		return e.funcToNumber(data)
	case "ltrimstr":
		return e.funcLTrimStr(argsStr, data)
	case "rtrimstr":
		return e.funcRTrimStr(argsStr, data)
	case "to_entries":
		return e.funcToEntries(data)
	case "from_entries":
		return e.funcFromEntries(data)
	case "with_entries":
		return e.funcWithEntries(argsStr, data)
	default:
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}
}

// funcLength returns the length of arrays, objects, strings, or null
func (e *Engine) funcLength(data interface{}) (interface{}, error) {
	if data == nil {
		return 0, nil
	}

	switch v := data.(type) {
	case []interface{}:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	case string:
		return len(v), nil
	default:
		return nil, fmt.Errorf("length not supported for type %T", data)
	}
}

// funcKeys returns the keys of an object or indices of an array
func (e *Engine) funcKeys(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		keys := make([]interface{}, 0, len(v))
		// Sort keys for deterministic output
		sortedKeys := make([]string, 0, len(v))
		for k := range v {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		for _, k := range sortedKeys {
			keys = append(keys, k)
		}
		return keys, nil
	case []interface{}:
		// Return array indices
		indices := make([]interface{}, len(v))
		for i := range v {
			indices[i] = i
		}
		return indices, nil
	default:
		return nil, fmt.Errorf("keys not supported for type %T", data)
	}
}

// funcValues returns the values of an object or array
func (e *Engine) funcValues(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Sort by keys for deterministic output
		sortedKeys := make([]string, 0, len(v))
		for k := range v {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		values := make([]interface{}, 0, len(v))
		for _, k := range sortedKeys {
			values = append(values, v[k])
		}
		return values, nil
	case []interface{}:
		// For arrays, values is the array itself
		return v, nil
	default:
		return nil, fmt.Errorf("values not supported for type %T", data)
	}
}

// funcType returns the type of the value
func (e *Engine) funcType(data interface{}) (interface{}, error) {
	if data == nil {
		return "null", nil
	}

	switch data.(type) {
	case bool:
		return "boolean", nil
	case float64, int, int64:
		return "number", nil
	case string:
		return "string", nil
	case []interface{}:
		return "array", nil
	case map[string]interface{}:
		return "object", nil
	default:
		return fmt.Sprintf("unknown(%T)", data), nil
	}
}

// funcSort sorts an array
func (e *Engine) funcSort(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("sort requires an array")
	}

	// Create a copy to avoid modifying original
	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	// Sort based on type
	sort.Slice(sorted, func(i, j int) bool {
		return compareForSort(sorted[i], sorted[j]) < 0
	})

	return sorted, nil
}

// funcReverse reverses an array
func (e *Engine) funcReverse(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("reverse requires an array")
	}

	reversed := make([]interface{}, len(arr))
	for i, v := range arr {
		reversed[len(arr)-1-i] = v
	}

	return reversed, nil
}

// compareForSort compares two values for sorting
func compareForSort(a, b interface{}) int {
	// Handle nil
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Try numeric comparison
	aNum, aOk := toNumber(a)
	bNum, bOk := toNumber(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	// String comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}

// funcMap applies an expression to each element of an array
func (e *Engine) funcMap(expr string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("map requires an array")
	}

	result := make([]interface{}, len(arr))
	for i, elem := range arr {
		mapped, err := e.executeQuery(strings.TrimSpace(expr), elem)
		if err != nil {
			return nil, fmt.Errorf("map error at index %d: %w", i, err)
		}
		result[i] = mapped
	}

	return result, nil
}

// funcSortBy sorts an array by the result of an expression
func (e *Engine) funcSortBy(expr string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("sort_by requires an array")
	}

	// Create a copy with computed sort keys
	type sortItem struct {
		value   interface{}
		sortKey interface{}
	}

	items := make([]sortItem, len(arr))
	for i, elem := range arr {
		sortKey, err := e.executeQuery(strings.TrimSpace(expr), elem)
		if err != nil {
			return nil, fmt.Errorf("sort_by error at index %d: %w", i, err)
		}
		items[i] = sortItem{value: elem, sortKey: sortKey}
	}

	// Sort by the computed keys
	sort.Slice(items, func(i, j int) bool {
		return compareForSort(items[i].sortKey, items[j].sortKey) < 0
	})

	// Extract sorted values
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = item.value
	}

	return result, nil
}

// funcGroupBy groups array elements by the result of an expression
func (e *Engine) funcGroupBy(expr string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("group_by requires an array")
	}

	// Group by computed keys
	groups := make(map[string][]interface{})
	for i, elem := range arr {
		groupKey, err := e.executeQuery(strings.TrimSpace(expr), elem)
		if err != nil {
			return nil, fmt.Errorf("group_by error at index %d: %w", i, err)
		}

		keyStr := fmt.Sprintf("%v", groupKey)
		groups[keyStr] = append(groups[keyStr], elem)
	}

	// Convert to array of arrays
	result := make([]interface{}, 0, len(groups))
	// Sort keys for deterministic output
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, groups[k])
	}

	return result, nil
}

// funcHas checks if an object has a given key
func (e *Engine) funcHas(key string, data interface{}) (interface{}, error) {
	key = strings.TrimSpace(key)
	// Remove quotes if present
	key = strings.Trim(key, `"`)

	obj, ok := data.(map[string]interface{})
	if !ok {
		return false, nil
	}

	_, exists := obj[key]
	return exists, nil
}

// funcIn checks if a value exists in an object's values or array
func (e *Engine) funcIn(container string, data interface{}) (interface{}, error) {
	// Execute the container expression
	containerData, err := e.executeQuery(strings.TrimSpace(container), data)
	if err != nil {
		return nil, err
	}

	switch c := containerData.(type) {
	case []interface{}:
		// Check if data is in array
		for _, v := range c {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", data) {
				return true, nil
			}
		}
		return false, nil
	case map[string]interface{}:
		// Check if data is in object values
		for _, v := range c {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", data) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

// funcSplit splits a string by a delimiter
func (e *Engine) funcSplit(delimiter string, data interface{}) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("split requires a string")
	}

	delimiter = strings.TrimSpace(delimiter)
	delimiter = strings.Trim(delimiter, `"`)

	parts := strings.Split(str, delimiter)
	result := make([]interface{}, len(parts))
	for i, part := range parts {
		result[i] = part
	}

	return result, nil
}

// funcJoin joins an array of strings with a delimiter
func (e *Engine) funcJoin(delimiter string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("join requires an array")
	}

	delimiter = strings.TrimSpace(delimiter)
	delimiter = strings.Trim(delimiter, `"`)

	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = fmt.Sprintf("%v", v)
	}

	return strings.Join(parts, delimiter), nil
}

// funcStartsWith checks if a string starts with a prefix
func (e *Engine) funcStartsWith(prefix string, data interface{}) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return false, nil
	}

	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, `"`)

	return strings.HasPrefix(str, prefix), nil
}

// funcEndsWith checks if a string ends with a suffix
func (e *Engine) funcEndsWith(suffix string, data interface{}) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return false, nil
	}

	suffix = strings.TrimSpace(suffix)
	suffix = strings.Trim(suffix, `"`)

	return strings.HasSuffix(str, suffix), nil
}

// funcContains checks if a string contains a substring
func (e *Engine) funcContains(substring string, data interface{}) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return false, nil
	}

	substring = strings.TrimSpace(substring)
	substring = strings.Trim(substring, `"`)

	return strings.Contains(str, substring), nil
}

// funcAdd sums all numbers in an array
func (e *Engine) funcAdd(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("add requires an array")
	}

	var sum float64
	for i, v := range arr {
		num, ok := toNumber(v)
		if !ok {
			return nil, fmt.Errorf("add: element at index %d is not a number", i)
		}
		sum += num
	}

	return sum, nil
}

// funcMin returns the minimum value from an array
func (e *Engine) funcMin(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("min requires an array")
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("min: empty array")
	}

	minNum, ok := toNumber(arr[0])
	if !ok {
		return nil, fmt.Errorf("min: first element is not a number")
	}

	for i := 1; i < len(arr); i++ {
		num, ok := toNumber(arr[i])
		if !ok {
			return nil, fmt.Errorf("min: element at index %d is not a number", i)
		}
		if num < minNum {
			minNum = num
		}
	}

	return minNum, nil
}

// funcMax returns the maximum value from an array
func (e *Engine) funcMax(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("max requires an array")
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("max: empty array")
	}

	maxNum, ok := toNumber(arr[0])
	if !ok {
		return nil, fmt.Errorf("max: first element is not a number")
	}

	for i := 1; i < len(arr); i++ {
		num, ok := toNumber(arr[i])
		if !ok {
			return nil, fmt.Errorf("max: element at index %d is not a number", i)
		}
		if num > maxNum {
			maxNum = num
		}
	}

	return maxNum, nil
}

// funcFloor returns the floor of a number
func (e *Engine) funcFloor(data interface{}) (interface{}, error) {
	num, ok := toNumber(data)
	if !ok {
		return nil, fmt.Errorf("floor requires a number")
	}

	return math.Floor(num), nil
}

// funcCeil returns the ceiling of a number
func (e *Engine) funcCeil(data interface{}) (interface{}, error) {
	num, ok := toNumber(data)
	if !ok {
		return nil, fmt.Errorf("ceil requires a number")
	}

	return math.Ceil(num), nil
}

// funcRound rounds a number to the nearest integer
func (e *Engine) funcRound(data interface{}) (interface{}, error) {
	num, ok := toNumber(data)
	if !ok {
		return nil, fmt.Errorf("round requires a number")
	}

	return math.Round(num), nil
}

// funcUnique returns unique elements from an array
func (e *Engine) funcUnique(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unique requires an array")
	}

	seen := make(map[string]bool)
	result := make([]interface{}, 0)

	for _, v := range arr {
		// Use string representation as key
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	return result, nil
}

// funcFlatten flattens an array (optionally to a specified depth)
func (e *Engine) funcFlatten(depthStr string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("flatten requires an array")
	}

	// Parse depth (default to 1 level if not specified)
	depth := 1
	if depthStr != "" {
		var err error
		depth, err = strconv.Atoi(strings.TrimSpace(depthStr))
		if err != nil {
			return nil, fmt.Errorf("flatten: invalid depth: %w", err)
		}
	}

	return flattenArray(arr, depth), nil
}

// flattenArray recursively flattens an array to the specified depth
func flattenArray(arr []interface{}, depth int) []interface{} {
	if depth <= 0 {
		return arr
	}

	result := make([]interface{}, 0)
	for _, v := range arr {
		if subArr, ok := v.([]interface{}); ok {
			// Recursively flatten sub-arrays
			flattened := flattenArray(subArr, depth-1)
			result = append(result, flattened...)
		} else {
			result = append(result, v)
		}
	}

	return result
}

// funcRange generates a range of numbers
func (e *Engine) funcRange(argsStr string, data interface{}) (interface{}, error) {
	// Parse arguments: range(n) or range(from; to) or range(from; to; step)
	args := strings.Split(argsStr, ";")

	var from, to, step int
	var err error

	switch len(args) {
	case 0, 1:
		// range(n) - from 0 to n-1
		if argsStr == "" {
			return nil, fmt.Errorf("range requires at least one argument")
		}
		to, err = strconv.Atoi(strings.TrimSpace(argsStr))
		if err != nil {
			return nil, fmt.Errorf("range: invalid argument: %w", err)
		}
		from = 0
		step = 1
	case 2:
		// range(from; to)
		from, err = strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil {
			return nil, fmt.Errorf("range: invalid 'from': %w", err)
		}
		to, err = strconv.Atoi(strings.TrimSpace(args[1]))
		if err != nil {
			return nil, fmt.Errorf("range: invalid 'to': %w", err)
		}
		step = 1
	case 3:
		// range(from; to; step)
		from, err = strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil {
			return nil, fmt.Errorf("range: invalid 'from': %w", err)
		}
		to, err = strconv.Atoi(strings.TrimSpace(args[1]))
		if err != nil {
			return nil, fmt.Errorf("range: invalid 'to': %w", err)
		}
		step, err = strconv.Atoi(strings.TrimSpace(args[2]))
		if err != nil {
			return nil, fmt.Errorf("range: invalid 'step': %w", err)
		}
		if step == 0 {
			return nil, fmt.Errorf("range: step cannot be zero")
		}
	default:
		return nil, fmt.Errorf("range: too many arguments")
	}

	// Generate range
	result := make([]interface{}, 0)
	if step > 0 {
		for i := from; i < to; i += step {
			result = append(result, i)
		}
	} else {
		for i := from; i > to; i += step {
			result = append(result, i)
		}
	}

	return result, nil
}

// funcFirst returns the first element(s) of an array
func (e *Engine) funcFirst(argsStr string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("first requires an array")
	}

	if len(arr) == 0 {
		return nil, nil
	}

	// If no argument, return first element
	if argsStr == "" {
		return arr[0], nil
	}

	// Otherwise return first n elements
	n, err := strconv.Atoi(strings.TrimSpace(argsStr))
	if err != nil {
		return nil, fmt.Errorf("first: invalid argument: %w", err)
	}

	if n < 0 {
		return nil, fmt.Errorf("first: argument must be non-negative")
	}

	if n > len(arr) {
		n = len(arr)
	}

	return arr[:n], nil
}

// funcLast returns the last element(s) of an array
func (e *Engine) funcLast(argsStr string, data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("last requires an array")
	}

	if len(arr) == 0 {
		return nil, nil
	}

	// If no argument, return last element
	if argsStr == "" {
		return arr[len(arr)-1], nil
	}

	// Otherwise return last n elements
	n, err := strconv.Atoi(strings.TrimSpace(argsStr))
	if err != nil {
		return nil, fmt.Errorf("last: invalid argument: %w", err)
	}

	if n < 0 {
		return nil, fmt.Errorf("last: argument must be non-negative")
	}

	if n > len(arr) {
		n = len(arr)
	}

	return arr[len(arr)-n:], nil
}

// funcToString converts a value to its string representation
func (e *Engine) funcToString(data interface{}) (interface{}, error) {
	if data == nil {
		return "null", nil
	}

	switch v := data.(type) {
	case string:
		return v, nil
	case float64:
		// Format numbers cleanly (avoid scientific notation for integers)
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v)), nil
		}
		return fmt.Sprintf("%g", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case []interface{}, map[string]interface{}:
		return nil, fmt.Errorf("tostring cannot convert arrays or objects")
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// funcToNumber converts a string to a number
func (e *Engine) funcToNumber(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case float64:
		return v, nil
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("tonumber: cannot parse '%s' as number: %w", v, err)
		}
		return num, nil
	case bool:
		if v {
			return float64(1), nil
		}
		return float64(0), nil
	case nil:
		return nil, fmt.Errorf("tonumber: cannot convert null to number")
	default:
		return nil, fmt.Errorf("tonumber: cannot convert %T to number", data)
	}
}

// funcLTrimStr removes a prefix string from the input
func (e *Engine) funcLTrimStr(argsStr string, data interface{}) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("ltrimstr requires a string, got %T", data)
	}

	// Parse the prefix argument
	prefix := strings.Trim(argsStr, "\"'")
	return strings.TrimPrefix(str, prefix), nil
}

// funcRTrimStr removes a suffix string from the input
func (e *Engine) funcRTrimStr(argsStr string, data interface{}) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("rtrimstr requires a string, got %T", data)
	}

	// Parse the suffix argument
	suffix := strings.Trim(argsStr, "\"'")
	return strings.TrimSuffix(str, suffix), nil
}

// funcToEntries converts an object to an array of {key, value} pairs
func (e *Engine) funcToEntries(data interface{}) (interface{}, error) {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("to_entries requires an object, got %T", data)
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build array of {key, value} objects
	entries := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		entry := map[string]interface{}{
			"key":   k,
			"value": obj[k],
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// funcFromEntries converts an array of {key, value} pairs to an object
func (e *Engine) funcFromEntries(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("from_entries requires an array, got %T", data)
	}

	result := make(map[string]interface{})
	for i, item := range arr {
		entry, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("from_entries: element %d is not an object", i)
		}

		// Support both {key, value} and {name, value} formats
		var key string
		if k, hasKey := entry["key"]; hasKey {
			key, ok = k.(string)
			if !ok {
				return nil, fmt.Errorf("from_entries: element %d has non-string key", i)
			}
		} else if n, hasName := entry["name"]; hasName {
			key, ok = n.(string)
			if !ok {
				return nil, fmt.Errorf("from_entries: element %d has non-string name", i)
			}
		} else {
			return nil, fmt.Errorf("from_entries: element %d missing 'key' or 'name' field", i)
		}

		value, hasValue := entry["value"]
		if !hasValue {
			return nil, fmt.Errorf("from_entries: element %d missing 'value' field", i)
		}

		result[key] = value
	}

	return result, nil
}

// funcWithEntries transforms object entries using an expression
func (e *Engine) funcWithEntries(argsStr string, data interface{}) (interface{}, error) {
	// First convert to entries
	entries, err := e.funcToEntries(data)
	if err != nil {
		return nil, err
	}

	// Apply the expression to each entry
	arr := entries.([]interface{})
	results := make([]interface{}, 0, len(arr))
	for _, entry := range arr {
		result, err := e.executeQuery(argsStr, entry)
		if err != nil {
			return nil, fmt.Errorf("with_entries: %w", err)
		}
		results = append(results, result)
	}

	// Convert back from entries
	return e.funcFromEntries(results)
}

func (e *Engine) executeIf(query string, data interface{}) (interface{}, error) {
	// Format: if COND then TRUE_BRANCH else FALSE_BRANCH end
	if !strings.HasSuffix(query, " end") {
		return nil, fmt.Errorf("if statement must end with 'end'")
	}

	// Remove "if " and " end"
	inner := query[3 : len(query)-4]

	// Find " then "
	thenIdx := strings.Index(inner, " then ")
	if thenIdx == -1 {
		return nil, fmt.Errorf("if statement missing 'then'")
	}

	condStr := strings.TrimSpace(inner[:thenIdx])
	rest := inner[thenIdx+6:]

	// Find " else "
	// Note: This is a simple implementation and might fail with nested if-else
	// A proper parser would be needed for nested structures
	elseIdx := strings.LastIndex(rest, " else ")
	if elseIdx == -1 {
		return nil, fmt.Errorf("if statement missing 'else'")
	}

	trueBranch := strings.TrimSpace(rest[:elseIdx])
	falseBranch := strings.TrimSpace(rest[elseIdx+6:])

	// Evaluate condition
	// We use evaluateCondition if it looks like a condition, otherwise executeQuery
	// In jq, any value can be a condition (null/false are false, others true)
	// For now, let's try to use evaluateCondition if it has operators, otherwise check truthiness
	var isTrue bool
	if strings.ContainsAny(condStr, "><=") {
		var err error
		isTrue, err = e.evaluateCondition(condStr, data)
		if err != nil {
			return nil, err
		}
	} else {
		val, err := e.executeQuery(condStr, data)
		if err != nil {
			return nil, err
		}
		isTrue = isTruthy(val)
	}

	if isTrue {
		return e.executeQuery(trueBranch, data)
	}
	return e.executeQuery(falseBranch, data)
}

func (e *Engine) executeAlternative(query string, data interface{}) (interface{}, error) {
	// Split by // respecting nesting
	parts := splitByString(query, "//")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		result, err := e.executeQuery(part, data)

		// If no error and result is truthy, return it
		if err == nil && isTruthy(result) {
			return result, nil
		}
		// If error, continue to next alternative?
		// jq behavior: errors in alternatives propagate, but null/false trigger next
		// For now, let's propagate errors
		if err != nil {
			return nil, err
		}
	}

	// If all alternatives are false/null, return the last one (or null/false)
	// Actually jq returns the last result if all are false/null
	// But we need to re-execute the last one to get the value?
	// We already executed it in the loop.
	// Wait, if we are here, it means the last one was also false/null (or empty parts)

	if len(parts) > 0 {
		// Re-execute last part to return its value
		return e.executeQuery(strings.TrimSpace(parts[len(parts)-1]), data)
	}

	return nil, nil
}

func isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return true
}

// splitByString splits a string by a separator, respecting nested structures.
func splitByString(s, sep string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	sepLen := len(sep)

	for i := 0; i < len(s); i++ {
		// Check for separator
		if depth == 0 && i+sepLen <= len(s) && s[i:i+sepLen] == sep {
			parts = append(parts, current.String())
			current.Reset()
			i += sepLen - 1 // Skip separator
			continue
		}

		ch := rune(s[i])
		switch ch {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		}
		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
