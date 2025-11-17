package toon

import (
	"testing"
)

func TestDecodeEmpty(t *testing.T) {
	_, err := Decode("")
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
}

func TestDecodeTabularArrayWithCommas(t *testing.T) {
	input := `items[2]{id,name,description}:
  1,Product A,"A product, with comma"
  2,Product B,"Another product, also with comma"`

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", result)
	}

	items, ok := obj["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items array, got %T", obj["items"])
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// Check first item
	item1, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map for item1, got %T", items[0])
	}

	if item1["description"] != "A product, with comma" {
		t.Errorf("Expected 'A product, with comma', got %v", item1["description"])
	}
}

func TestDecodeInvalidArrayLength(t *testing.T) {
	input := `items[-1]{id}: 1`

	_, err := Decode(input)
	if err == nil {
		t.Error("Expected error for negative array length, got nil")
	}
}

func TestDecodeNestedObject(t *testing.T) {
	input := `user:
  name: Alice
  age: 30`

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", result)
	}

	user, ok := obj["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user map, got %T", obj["user"])
	}

	if user["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", user["name"])
	}
}
