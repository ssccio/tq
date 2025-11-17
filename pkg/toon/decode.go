package toon

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Decode parses TOON format into a Go value
func Decode(input string) (interface{}, error) {
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}
	lines := strings.Split(input, "\n")
	return parseTOON(lines, 0)
}

func parseTOON(lines []string, startIdx int) (interface{}, error) {
	if startIdx >= len(lines) {
		return nil, nil
	}

	result := make(map[string]interface{})

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check indentation level
		indent := countIndent(line)

		// Parse key-value or array
		if strings.Contains(line, ":") {
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Check if it's an array
			if strings.HasPrefix(key, "[") || strings.Contains(key, "[") {
				// Parse array header
				arr, err := parseArrayHeader(key, value, lines, i+1)
				if err != nil {
					return nil, err
				}

				// Extract actual key if it has one
				actualKey := key
				if idx := strings.Index(key, "["); idx > 0 {
					actualKey = key[:idx]
				}

				result[actualKey] = arr
				continue
			}

			// Simple value
			if value != "" {
				parsed, err := parseValue(value)
				if err != nil {
					return nil, err
				}
				result[key] = parsed
			} else {
				// Nested object - parse following indented lines
				nested := make(map[string]interface{})
				j := i + 1
				for j < len(lines) && countIndent(lines[j]) > indent {
					j++
				}
				if j > i+1 {
					nestedResult, err := parseTOON(lines[i+1:j], 0)
					if err != nil {
						return nil, fmt.Errorf("failed to parse nested object at key '%s': %w", key, err)
					}
					// Validate type assertion
					nestedMap, ok := nestedResult.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("expected object for key '%s', got %T", key, nestedResult)
					}
					nested = nestedMap
				}
				result[key] = nested
			}
		}
	}

	return result, nil
}

func parseArrayHeader(key, value string, lines []string, nextIdx int) (interface{}, error) {
	// Extract array info: [length]{fields} or [length]
	var length int
	var fields []string

	// Parse [N]
	start := strings.Index(key, "[")
	end := strings.Index(key, "]")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("invalid array syntax in '%s'", key)
	}

	lengthStr := key[start+1 : end]
	if lengthStr == "" {
		return nil, fmt.Errorf("empty array length in '%s'", key)
	}

	var err error
	length, err = strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array length '%s': %w", lengthStr, err)
	}

	if length < 0 {
		return nil, fmt.Errorf("negative array length not allowed: %d", length)
	}

	// Check for fields {field1,field2}
	if strings.Contains(key, "{") {
		fieldStart := strings.Index(key, "{")
		fieldEnd := strings.Index(key, "}")
		if fieldStart == -1 || fieldEnd == -1 {
			return nil, fmt.Errorf("invalid array field syntax")
		}
		fieldStr := key[fieldStart+1 : fieldEnd]
		fields = strings.Split(fieldStr, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
	}

	// Parse array content
	if len(fields) > 0 {
		// Tabular array
		return parseTabularArray(length, fields, lines, nextIdx)
	}

	// Primitive or mixed array
	if value != "" {
		// Inline primitive array
		return parsePrimitiveArray(value)
	}

	// List format array
	return parseListArray(length, lines, nextIdx)
}

func parsePrimitiveArray(value string) (interface{}, error) {
	values := strings.Split(value, ",")
	result := make([]interface{}, 0, len(values))

	for _, v := range values {
		parsed, err := parseValue(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}

	return result, nil
}

func parseTabularArray(length int, fields []string, lines []string, startIdx int) (interface{}, error) {
	result := make([]interface{}, 0, length)

	for i := 0; i < length && startIdx+i < len(lines); i++ {
		line := strings.TrimSpace(lines[startIdx+i])
		if line == "" {
			continue
		}

		// Use CSV reader to properly handle quoted fields with commas
		csvReader := csv.NewReader(strings.NewReader(line))
		csvReader.Comma = ','
		csvReader.TrimLeadingSpace = true

		values, err := csvReader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to parse tabular row at line %d: %w", startIdx+i, err)
		}

		if len(values) != len(fields) {
			return nil, fmt.Errorf("field count mismatch at line %d: expected %d fields, got %d", startIdx+i, len(fields), len(values))
		}

		obj := make(map[string]interface{})
		for j, field := range fields {
			parsed, err := parseValue(strings.TrimSpace(values[j]))
			if err != nil {
				return nil, err
			}
			obj[field] = parsed
		}

		result = append(result, obj)
	}

	return result, nil
}

func parseListArray(length int, lines []string, startIdx int) (interface{}, error) {
	result := make([]interface{}, 0, length)

	for i := 0; i < length && startIdx+i < len(lines); i++ {
		line := strings.TrimSpace(lines[startIdx+i])
		if !strings.HasPrefix(line, "-") {
			continue
		}

		value := strings.TrimSpace(line[1:])
		parsed, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}

	return result, nil
}

func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	if s == "null" {
		return nil, nil
	}

	if s == "true" {
		return true, nil
	}

	if s == "false" {
		return false, nil
	}

	// Try number
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		// Check if it's an integer
		if float64(int64(num)) == num {
			return int64(num), nil
		}
		return num, nil
	}

	// String - remove quotes if present
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return strings.Trim(s, "\""), nil
	}

	return s, nil
}

func countIndent(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 2 // Treat tab as 2 spaces
		} else {
			break
		}
	}
	return count
}

// DecodeReader reads TOON from a reader
func DecodeReader(r *bufio.Reader) (interface{}, error) {
	var lines []string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					lines = append(lines, strings.TrimSuffix(line, "\n"))
				}
				break
			}
			return nil, err
		}
		lines = append(lines, strings.TrimSuffix(line, "\n"))
	}

	return parseTOON(lines, 0)
}
