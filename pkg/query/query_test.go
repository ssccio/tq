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

func TestExecuteArrayIterationWithField(t *testing.T) {
	engine := New()

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": float64(30)},
			map[string]interface{}{"name": "Bob", "age": float64(25)},
		},
	}

	// Test .users[].name
	result, err := engine.Execute(".users[].name", data)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected array, got %T", result)
	}

	if len(arr) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(arr))
	}

	if arr[0] != "Alice" || arr[1] != "Bob" {
		t.Errorf("Expected ['Alice', 'Bob'], got %v", arr)
	}
}

func TestExecuteArrayIterationWithSelect(t *testing.T) {
	engine := New()

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": float64(30)},
			map[string]interface{}{"name": "Bob", "age": float64(20)},
			map[string]interface{}{"name": "Charlie", "age": float64(35)},
		},
	}

	// Test .users[] | select(.age > 25)
	result, err := engine.Execute(".users[] | select(.age > 25)", data)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected array, got %T", result)
	}

	if len(arr) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(arr))
	}

	// Check first result
	first, ok := arr[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", arr[0])
	}
	if first["name"] != "Alice" {
		t.Errorf("Expected first result to be Alice, got %v", first["name"])
	}

	// Check second result
	second, ok := arr[1].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", arr[1])
	}
	if second["name"] != "Charlie" {
		t.Errorf("Expected second result to be Charlie, got %v", second["name"])
	}
}

func TestBuiltInFunctions(t *testing.T) {
	engine := New()

	// Test length
	t.Run("length", func(t *testing.T) {
		result, err := engine.Execute("length()", []interface{}{1, 2, 3})
		if err != nil {
			t.Fatalf("length() failed: %v", err)
		}
		if result != 3 {
			t.Errorf("Expected 3, got %v", result)
		}
	})

	// Test keys
	t.Run("keys", func(t *testing.T) {
		data := map[string]interface{}{"b": 2, "a": 1}
		result, err := engine.Execute("keys()", data)
		if err != nil {
			t.Fatalf("keys() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 2 || arr[0] != "a" || arr[1] != "b" {
			t.Errorf("Expected sorted keys [a, b], got %v", arr)
		}
	})

	// Test values
	t.Run("values", func(t *testing.T) {
		data := map[string]interface{}{"b": 2, "a": 1}
		result, err := engine.Execute("values()", data)
		if err != nil {
			t.Fatalf("values() failed: %v", err)
		}
		arr := result.([]interface{})
		// Values should be in key-sorted order
		if len(arr) != 2 {
			t.Errorf("Expected 2 values, got %d", len(arr))
		}
	})

	// Test type
	t.Run("type", func(t *testing.T) {
		result, err := engine.Execute("type()", []interface{}{1, 2})
		if err != nil {
			t.Fatalf("type() failed: %v", err)
		}
		if result != "array" {
			t.Errorf("Expected 'array', got %v", result)
		}
	})

	// Test sort
	t.Run("sort", func(t *testing.T) {
		data := []interface{}{3.0, 1.0, 2.0}
		result, err := engine.Execute("sort()", data)
		if err != nil {
			t.Fatalf("sort() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 3 || arr[0] != 1.0 || arr[1] != 2.0 || arr[2] != 3.0 {
			t.Errorf("Expected [1, 2, 3], got %v", arr)
		}
	})

	// Test map
	t.Run("map", func(t *testing.T) {
		data := []interface{}{
			map[string]interface{}{"x": float64(1)},
			map[string]interface{}{"x": float64(2)},
		}
		result, err := engine.Execute("map(.x)", data)
		if err != nil {
			t.Fatalf("map() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 2 || arr[0] != float64(1) || arr[1] != float64(2) {
			t.Errorf("Expected [1, 2], got %v", arr)
		}
	})

	// Test sort_by
	t.Run("sort_by", func(t *testing.T) {
		data := []interface{}{
			map[string]interface{}{"name": "Charlie"},
			map[string]interface{}{"name": "Alice"},
		}
		result, err := engine.Execute("sort_by(.name)", data)
		if err != nil {
			t.Fatalf("sort_by() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(arr))
		}
		first := arr[0].(map[string]interface{})
		if first["name"] != "Alice" {
			t.Errorf("Expected first to be Alice, got %v", first["name"])
		}
	})

	// Test has
	t.Run("has", func(t *testing.T) {
		data := map[string]interface{}{"name": "Alice"}
		result, err := engine.Execute(`has("name")`, data)
		if err != nil {
			t.Fatalf("has() failed: %v", err)
		}
		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})
}

func TestStringFunctions(t *testing.T) {
	engine := New()

	// Test split
	t.Run("split", func(t *testing.T) {
		result, err := engine.Execute(`split(",")`, "a,b,c")
		if err != nil {
			t.Fatalf("split() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 3 || arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
			t.Errorf("Expected [a, b, c], got %v", arr)
		}
	})

	// Test join
	t.Run("join", func(t *testing.T) {
		data := []interface{}{"hello", "world"}
		result, err := engine.Execute(`join(" ")`, data)
		if err != nil {
			t.Fatalf("join() failed: %v", err)
		}
		if result != "hello world" {
			t.Errorf("Expected 'hello world', got %v", result)
		}
	})

	// Test startswith
	t.Run("startswith", func(t *testing.T) {
		result, err := engine.Execute(`startswith("hello")`, "hello world")
		if err != nil {
			t.Fatalf("startswith() failed: %v", err)
		}
		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})

	// Test endswith
	t.Run("endswith", func(t *testing.T) {
		result, err := engine.Execute(`endswith("world")`, "hello world")
		if err != nil {
			t.Fatalf("endswith() failed: %v", err)
		}
		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})

	// Test contains
	t.Run("contains", func(t *testing.T) {
		result, err := engine.Execute(`contains("ll")`, "hello")
		if err != nil {
			t.Fatalf("contains() failed: %v", err)
		}
		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})
}

func TestMathFunctions(t *testing.T) {
	engine := New()

	// Test add
	t.Run("add", func(t *testing.T) {
		data := []interface{}{float64(1), float64(2), float64(3)}
		result, err := engine.Execute("add()", data)
		if err != nil {
			t.Fatalf("add() failed: %v", err)
		}
		if result != float64(6) {
			t.Errorf("Expected 6, got %v", result)
		}
	})

	// Test min
	t.Run("min", func(t *testing.T) {
		data := []interface{}{float64(3), float64(1), float64(2)}
		result, err := engine.Execute("min()", data)
		if err != nil {
			t.Fatalf("min() failed: %v", err)
		}
		if result != float64(1) {
			t.Errorf("Expected 1, got %v", result)
		}
	})

	// Test max
	t.Run("max", func(t *testing.T) {
		data := []interface{}{float64(3), float64(1), float64(2)}
		result, err := engine.Execute("max()", data)
		if err != nil {
			t.Fatalf("max() failed: %v", err)
		}
		if result != float64(3) {
			t.Errorf("Expected 3, got %v", result)
		}
	})

	// Test floor
	t.Run("floor", func(t *testing.T) {
		result, err := engine.Execute("floor()", float64(3.7))
		if err != nil {
			t.Fatalf("floor() failed: %v", err)
		}
		if result != float64(3) {
			t.Errorf("Expected 3, got %v", result)
		}
	})

	// Test ceil
	t.Run("ceil", func(t *testing.T) {
		result, err := engine.Execute("ceil()", float64(3.2))
		if err != nil {
			t.Fatalf("ceil() failed: %v", err)
		}
		if result != float64(4) {
			t.Errorf("Expected 4, got %v", result)
		}
	})

	// Test round
	t.Run("round", func(t *testing.T) {
		result, err := engine.Execute("round()", float64(3.7))
		if err != nil {
			t.Fatalf("round() failed: %v", err)
		}
		if result != float64(4) {
			t.Errorf("Expected 4, got %v", result)
		}
	})
}

func TestArrayUtilityFunctions(t *testing.T) {
	engine := New()

	// Test unique
	t.Run("unique", func(t *testing.T) {
		data := []interface{}{float64(1), float64(2), float64(2), float64(3), float64(1)}
		result, err := engine.Execute("unique()", data)
		if err != nil {
			t.Fatalf("unique() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 3 {
			t.Errorf("Expected 3 unique elements, got %d", len(arr))
		}
		// Should maintain order of first occurrence
		if arr[0] != float64(1) || arr[1] != float64(2) || arr[2] != float64(3) {
			t.Errorf("Expected [1, 2, 3], got %v", arr)
		}
	})

	// Test flatten with default depth
	t.Run("flatten", func(t *testing.T) {
		data := []interface{}{
			[]interface{}{float64(1), float64(2)},
			[]interface{}{float64(3), float64(4)},
		}
		result, err := engine.Execute("flatten()", data)
		if err != nil {
			t.Fatalf("flatten() failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 4 {
			t.Fatalf("Expected 4 elements, got %d", len(arr))
		}
		if arr[0] != float64(1) || arr[1] != float64(2) || arr[2] != float64(3) || arr[3] != float64(4) {
			t.Errorf("Expected [1, 2, 3, 4], got %v", arr)
		}
	})

	// Test flatten with depth parameter
	t.Run("flatten_depth", func(t *testing.T) {
		data := []interface{}{
			[]interface{}{
				[]interface{}{float64(1), float64(2)},
			},
			[]interface{}{
				[]interface{}{float64(3), float64(4)},
			},
		}
		result, err := engine.Execute("flatten(2)", data)
		if err != nil {
			t.Fatalf("flatten(2) failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 4 {
			t.Fatalf("Expected 4 elements after flattening depth 2, got %d", len(arr))
		}
	})

	// Test range with single argument
	t.Run("range_single", func(t *testing.T) {
		result, err := engine.Execute("range(5)", nil)
		if err != nil {
			t.Fatalf("range(5) failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 5 {
			t.Fatalf("Expected 5 elements, got %d", len(arr))
		}
		for i := 0; i < 5; i++ {
			if arr[i] != i {
				t.Errorf("Expected element %d to be %d, got %v", i, i, arr[i])
			}
		}
	})

	// Test range with start and end
	t.Run("range_start_end", func(t *testing.T) {
		result, err := engine.Execute("range(2;5)", nil)
		if err != nil {
			t.Fatalf("range(2;5) failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 3 {
			t.Fatalf("Expected 3 elements, got %d", len(arr))
		}
		if arr[0] != 2 || arr[1] != 3 || arr[2] != 4 {
			t.Errorf("Expected [2, 3, 4], got %v", arr)
		}
	})

	// Test range with step
	t.Run("range_step", func(t *testing.T) {
		result, err := engine.Execute("range(0;10;2)", nil)
		if err != nil {
			t.Fatalf("range(0;10;2) failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 5 {
			t.Fatalf("Expected 5 elements, got %d", len(arr))
		}
		if arr[0] != 0 || arr[1] != 2 || arr[2] != 4 || arr[3] != 6 || arr[4] != 8 {
			t.Errorf("Expected [0, 2, 4, 6, 8], got %v", arr)
		}
	})

	// Test first without argument
	t.Run("first", func(t *testing.T) {
		data := []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}
		result, err := engine.Execute("first()", data)
		if err != nil {
			t.Fatalf("first() failed: %v", err)
		}
		if result != float64(1) {
			t.Errorf("Expected 1, got %v", result)
		}
	})

	// Test first with count argument
	t.Run("first_n", func(t *testing.T) {
		data := []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}
		result, err := engine.Execute("first(3)", data)
		if err != nil {
			t.Fatalf("first(3) failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 3 {
			t.Fatalf("Expected 3 elements, got %d", len(arr))
		}
		if arr[0] != float64(1) || arr[1] != float64(2) || arr[2] != float64(3) {
			t.Errorf("Expected [1, 2, 3], got %v", arr)
		}
	})

	// Test last without argument
	t.Run("last", func(t *testing.T) {
		data := []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}
		result, err := engine.Execute("last()", data)
		if err != nil {
			t.Fatalf("last() failed: %v", err)
		}
		if result != float64(5) {
			t.Errorf("Expected 5, got %v", result)
		}
	})

	// Test last with count argument
	t.Run("last_n", func(t *testing.T) {
		data := []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}
		result, err := engine.Execute("last(2)", data)
		if err != nil {
			t.Fatalf("last(2) failed: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 2 {
			t.Fatalf("Expected 2 elements, got %d", len(arr))
		}
		if arr[0] != float64(4) || arr[1] != float64(5) {
			t.Errorf("Expected [4, 5], got %v", arr)
		}
	})
}

func TestExecuteIf(t *testing.T) {
	engine := New()

	t.Run("basic_true", func(t *testing.T) {
		result, err := engine.Execute("if true then 1 else 2 end", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != float64(1) {
			t.Errorf("Expected 1, got %v", result)
		}
	})

	t.Run("basic_false", func(t *testing.T) {
		result, err := engine.Execute("if false then 1 else 2 end", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != float64(2) {
			t.Errorf("Expected 2, got %v", result)
		}
	})

	t.Run("condition_expression", func(t *testing.T) {
		data := map[string]interface{}{"age": float64(30)}
		result, err := engine.Execute("if .age > 25 then \"adult\" else \"young\" end", data)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "adult" {
			t.Errorf("Expected 'adult', got %v", result)
		}
	})

	t.Run("truthy_check", func(t *testing.T) {
		// Non-null/non-false is true
		result, err := engine.Execute("if 123 then \"yes\" else \"no\" end", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "yes" {
			t.Errorf("Expected 'yes', got %v", result)
		}
	})

	t.Run("falsy_check_null", func(t *testing.T) {
		// null is false
		data := map[string]interface{}{"val": nil}
		result, err := engine.Execute("if .val then \"yes\" else \"no\" end", data)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "no" {
			t.Errorf("Expected 'no', got %v", result)
		}
	})
}

func TestExecuteAlternative(t *testing.T) {
	engine := New()

	t.Run("first_truthy", func(t *testing.T) {
		result, err := engine.Execute("1 // 2", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != float64(1) {
			t.Errorf("Expected 1, got %v", result)
		}
	})

	t.Run("second_truthy", func(t *testing.T) {
		result, err := engine.Execute("false // 2", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != float64(2) {
			t.Errorf("Expected 2, got %v", result)
		}
	})

	t.Run("null_fallback", func(t *testing.T) {
		data := map[string]interface{}{"a": nil, "b": float64(3)}
		result, err := engine.Execute(".a // .b", data)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != float64(3) {
			t.Errorf("Expected 3, got %v", result)
		}
	})

	t.Run("chain", func(t *testing.T) {
		result, err := engine.Execute("false // null // 3", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != float64(3) {
			t.Errorf("Expected 3, got %v", result)
		}
	})

	t.Run("all_falsy", func(t *testing.T) {
		result, err := engine.Execute("false // null", nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})
}
