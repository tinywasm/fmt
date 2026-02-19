package fmt

import "testing"

// TestConvertVariadicValidation tests Convert() parameter validation
func TestConvertVariadicValidation(t *testing.T) {
	// Valid usage
	c1 := Convert()        // Empty - should work
	c2 := Convert("hello") // Single value - should work
	if len(c1.err) > 0 {
		t.Errorf("Convert() should not have error, got: %s", c1.getError())
	}
	if len(c2.err) > 0 {
		t.Errorf("Convert(value) should not have error, got: %s", c2.getError())
	}

	// Clean up
	c1.putConv()
	c2.putConv()
	// Invalid usage - should set error and continue chain
	c3 := Convert("hello", "world") // Multiple values - should set error
	if len(c3.err) == 0 {
		t.Error("Convert with multiple parameters should set error")
	}

	// Chain should continue but operations should be omitted due to error
	out := c3.Write(" more").String() // This auto-releases
	if out != "" {
		t.Errorf("Operations after error should be omitted, got: %s", out)
	}
}

// TestWriteMethod tests the unified Write method
func TestWriteMethod(t *testing.T) {
	tests := []struct {
		name     string
		values   []any
		expected string
	}{
		{"String values", []any{"hello", " ", "world"}, "hello world"},
		{"Mixed types", []any{"Count: ", 42, " items"}, "Count: 42 items"},
		{"Boolean values", []any{"Active: ", true, ", Valid: ", false}, "Active: true, Valid: false"},
		{"Float values", []any{"Price: $", 19.99}, "Price: $19.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Convert()
			for _, v := range tt.values {
				c.Write(v)
			}
			out := c.String() // Auto-releases

			if out != tt.expected {
				t.Errorf("Write chain failed: got %q, want %q", out, tt.expected)
			}
		})
	}
}

// TestResetMethod tests the Reset functionality
func TestResetMethod(t *testing.T) {
	c := Convert("initial")
	c.Write(" more")

	// Reset and reuse
	c.Reset()
	c.Write("new").Write(" content")
	out := c.String() // Auto-releases

	expected := "new content"
	if out != expected {
		t.Errorf("Reset failed: got %q, want %q", out, expected)
	}
}

// TestErrorChainInterruption tests error chain interruption behavior
func TestErrorChainInterruption(t *testing.T) {
	// Test normal case first
	c := Convert("valid")
	c.Write("ok")
	out := c.String() // Auto-releases
	expected := "validok"
	if out != expected {
		t.Errorf("Normal chain failed: got %q, want %q", out, expected)
	}
	// Test error case
	c2 := Convert("hello", "world") // This should set error
	if len(c2.err) == 0 {
		t.Error("Expected error for multiple parameters, got none")
	}

	c2.Write(" more") // This should be omitted due to error

	result2, err := c2.StringErr()
	if err == nil {
		t.Error("Expected error from StringErr(), got nil")
	}
	// When there's an error, out should be empty string
	if result2 != "" {
		t.Errorf("Expected empty out due to error, got: %s", result2)
	}
}

// TestBuilderPattern tests the main optimization goal: empty Convert() for loops
func TestBuilderPattern(t *testing.T) {
	items := []string{"  APPLE  ", "  banana  ", "  Cherry  "}

	// Test builder pattern with transformations
	// Process each item individually, then combine
	c := Convert() // Empty initialization
	for i, item := range items {
		// Process each item individually and append to builder
		processed := Convert(item).TrimSpace().ToLower().Capitalize().String()
		c.Write(processed)
		if i < len(items)-1 {
			c.Write(" - ")
		}
	}
	out := c.String() // Auto-releases

	expected := "Apple - Banana - Cherry"
	if out != expected {
		t.Errorf("Builder pattern failed: got %q, want %q", out, expected)
	}

	// Test simple pattern too
	c2 := Convert() // Empty initialization
	for _, item := range []string{"apple", "banana", "cherry"} {
		c2.Write(item).Write(" ")
	}
	result2 := c2.String() // Auto-releases

	expected2 := "apple banana cherry "
	if result2 != expected2 {
		t.Errorf("Simple builder pattern failed: got %q, want %q", result2, expected2)
	}
}

// TestTFunction tests the Translate translation function
func TestTFunction(t *testing.T) {
	RegisterWords([]DictEntry{
		{EN: "Blorp", ES: "Blorpes"},
		{EN: "Quux", ES: "Quuxes"},
	})
	out := Translate("blorp", "quux").String()
	if out == "" {
		t.Error("Translate function returned empty string")
	}

	result2 := Translate(ES, "blorp", "quux").String()
	if result2 == "" {
		t.Error("Translate function with language returned empty string")
	}

	if out == result2 {
		t.Errorf("Translate function should return different translations: %q vs %q", out, result2)
	}
}

// TestErrFunction tests the refactored Err function
func TestErrFunction(t *testing.T) {
	err := Err("rawfmt", "rawval")
	if len(err.err) == 0 {
		t.Error("Err function should create error message")
	}
}
