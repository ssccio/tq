package toon

import (
	"testing"
)

func TestEncodePrimitiveArray(t *testing.T) {
	opts := DefaultOptions()

	data := map[string]interface{}{
		"tags": []interface{}{"admin", "ops", "dev"},
	}

	result, err := Encode(data, opts)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := "tags[3]: admin,ops,dev"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestEncodeTabularArray(t *testing.T) {
	opts := DefaultOptions()

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "Alice", "role": "admin"},
			map[string]interface{}{"id": float64(2), "name": "Bob", "role": "user"},
		},
	}

	result, err := Encode(data, opts)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Check it contains the header and data
	if !contains(result, "[2]{") {
		t.Errorf("Expected tabular array header, got:\n%s", result)
	}
}

func TestEncodeNestedObject(t *testing.T) {
	opts := DefaultOptions()

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   float64(123),
			"name": "Ada",
		},
	}

	result, err := Encode(data, opts)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !contains(result, "user:") {
		t.Errorf("Expected nested object, got:\n%s", result)
	}
}

func TestEncodeString(t *testing.T) {
	tests := []struct {
		input     string
		delimiter string
		wantQuote bool
	}{
		{"normal", ",", false},
		{"hello world", ",", false},
		{" padded ", ",", true},
		{"true", ",", true},
		{"42", ",", true},
		{"a,b", ",", true},
		{"a:b", ",", true},
		{"- item", ",", true},
	}

	for _, tt := range tests {
		result := encodeString(tt.input, tt.delimiter)
		hasQuote := result[0] == '"'

		if hasQuote != tt.wantQuote {
			t.Errorf("encodeString(%q) = %q, want quote=%v", tt.input, result, tt.wantQuote)
		}
	}
}

func TestDecodeSimpleObject(t *testing.T) {
	input := `id: 123
name: Ada
active: true`

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected object, got %T", result)
	}

	if obj["name"] != "Ada" {
		t.Errorf("Expected name=Ada, got %v", obj["name"])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
