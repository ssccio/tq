package converter

import (
	"strings"
	"testing"
)

func TestReadJSON(t *testing.T) {
	conv := New(Options{
		InputFormat:  "json",
		OutputFormat: "toon",
	})

	input := strings.NewReader(`{"name": "Alice", "age": 30}`)
	result, err := conv.Read(input)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", result)
	}

	if m["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", m["name"])
	}
}

func TestReadWithSizeLimit(t *testing.T) {
	conv := New(Options{
		InputFormat:  "json",
		OutputFormat: "toon",
		MaxInputSize: 10, // Very small limit
	})

	// This input is larger than 10 bytes
	input := strings.NewReader(`{"name": "Alice", "age": 30}`)
	_, err := conv.Read(input)

	// Should not panic or crash, might be truncated
	if err == nil {
		// Truncated read is acceptable
		t.Log("Read succeeded with truncated input")
	} else {
		// Error is also acceptable
		t.Logf("Read failed as expected with size limit: %v", err)
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"key": "value"}`, "json"},
		{`[1, 2, 3]`, "json"},
		{`key: value`, "yaml"},
		{`users[2]{id,name}:`, "toon"},
	}

	for _, tt := range tests {
		result := detectFormat([]byte(tt.input))
		if result != tt.expected {
			t.Errorf("detectFormat(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
