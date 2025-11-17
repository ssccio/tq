package query

import (
	"testing"
)

func TestExecuteArrayIndexChained(t *testing.T) {
	engine := New()

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": float64(30)},
			map[string]interface{}{"name": "Bob", "age": float64(25)},
		},
	}

	// Test chained array access
	result, err := engine.Execute(".users[0].name", data)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != "Alice" {
		t.Errorf("Expected 'Alice', got %v", result)
	}
}

func TestExecuteArrayIndexNegative(t *testing.T) {
	engine := New()

	data := []interface{}{"a", "b", "c"}

	result, err := engine.Execute(".[-1]", data)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != "c" {
		t.Errorf("Expected 'c', got %v", result)
	}
}

func TestExecuteFieldAccess(t *testing.T) {
	engine := New()

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  float64(30),
		},
	}

	result, err := engine.Execute(".user.name", data)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != "Alice" {
		t.Errorf("Expected 'Alice', got %v", result)
	}
}

func TestExecuteIdentity(t *testing.T) {
	engine := New()

	data := map[string]interface{}{
		"test": "value",
	}

	result, err := engine.Execute(".", data)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", result)
	}

	if resultMap["test"] != "value" {
		t.Errorf("Expected 'value', got %v", resultMap["test"])
	}
}
