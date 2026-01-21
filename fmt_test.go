package fmt

import "testing"

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "Fmt integer with thousand separators (EU)",
			input: 2189009,
			want:  "2.189.009",
		},
		{
			name:  "Fmt decimal number with trailing zeros (EU)",
			input: 2189009.00,
			want:  "2.189.009",
		},
		{
			name:  "Fmt decimal number (EU)",
			input: 2189009.123,
			want:  "2.189.009,123",
		},
		{
			name:  "Fmt string number (EU)",
			input: "2189009.00",
			want:  "2.189.009",
		},
		{
			name:  "Fmt negative number (EU)",
			input: -2189009,
			want:  "-2.189.009",
		},
		{
			name:  "Fmt small number",
			input: 123,
			want:  "123",
		},
		{
			name:  "Fmt zero",
			input: 0,
			want:  "0",
		},
		{
			name:  "Non-numeric input",
			input: "hello",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := Convert(tt.input).Thousands().String()
			if out != tt.want {
				t.Errorf("Thousands() got = %v, want %v", out, tt.want)
			}
		})
	}

	// Anglo format tests
	t.Run("Fmt integer with thousand separators (Anglo)", func(t *testing.T) {
		out := Convert(2189009).Thousands(true).String()
		if out != "2,189,009" {
			t.Errorf("Thousands(true) got = %v, want %v", out, "2,189,009")
		}
	})
	t.Run("Fmt decimal number with trailing zeros (Anglo)", func(t *testing.T) {
		out := Convert(2189009.00).Thousands(true).String()
		if out != "2,189,009" {
			t.Errorf("Thousands(true) got = %v, want %v", out, "2,189,009")
		}
	})
	t.Run("Fmt decimal number (Anglo)", func(t *testing.T) {
		out := Convert(2189009.123).Thousands(true).String()
		if out != "2,189,009.123" {
			t.Errorf("Thousands(true) got = %v, want %v", out, "2,189,009.123")
		}
	})
	t.Run("Fmt string number (Anglo)", func(t *testing.T) {
		out := Convert("2189009.00").Thousands(true).String()
		if out != "2,189,009" {
			t.Errorf("Thousands(true) got = %v, want %v", out, "2,189,009")
		}
	})
	t.Run("Fmt negative number (Anglo)", func(t *testing.T) {
		out := Convert(-2189009).Thousands(true).String()
		if out != "-2,189,009" {
			t.Errorf("Thousands(true) got = %v, want %v", out, "-2,189,009")
		}
	})
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "String formatting",
			format:   "Hello %s!",
			args:     []any{"World"},
			expected: "Hello World!",
		},
		{
			name:     "Integer formatting",
			format:   "Value: %d",
			args:     []any{42},
			expected: "Value: 42",
		},
		{
			name:     "Float formatting",
			format:   "Pi: %.2f",
			args:     []any{3.14159},
			expected: "Pi: 3.14", // Changed to match Round default (ceiling)
		},
		{
			name:     "Multiple arguments",
			format:   "Hello %s, you have %d messages",
			args:     []any{"Alice", 5},
			expected: "Hello Alice, you have 5 messages",
		},
		{
			name:     "Binary formatting",
			format:   "Binary: %b",
			args:     []any{7},
			expected: "Binary: 111",
		},
		{
			name:     "Hexadecimal formatting",
			format:   "Hex: %x",
			args:     []any{255},
			expected: "Hex: ff",
		},
		{
			name:     "Octal formatting",
			format:   "Octal: %o",
			args:     []any{64},
			expected: "Octal: 100",
		},
		{
			name:     "Value formatting",
			format:   "Bool: %v",
			args:     []any{true},
			expected: "Bool: true",
		},
		{
			name:     "Percent sign",
			format:   "100%% complete",
			args:     []any{},
			expected: "100% complete",
		},
		{
			name:     "Missing argument",
			format:   "Value: %d",
			args:     []any{},
			expected: "",
		},
		// --- Especificadores estándar de Go no cubiertos ---
		{
			name:     "Unsigned integer formatting",
			format:   "Value: %u",
			args:     []any{uint(42)},
			expected: "Value: 42",
		},
		{
			name:     "Boolean true formatting",
			format:   "Bool: %t",
			args:     []any{true},
			expected: "Bool: true",
		},
		{
			name:     "Boolean false formatting",
			format:   "Bool: %t",
			args:     []any{false},
			expected: "Bool: false",
		},
		{
			name:     "Quoted string formatting",
			format:   "Quoted: %q",
			args:     []any{"hello"},
			expected: "Quoted: \"hello\"",
		},
		{
			name:     "Quoted char formatting",
			format:   "Quoted: %q",
			args:     []any{'A'},
			expected: "Quoted: 'A'",
		},
		{
			name:     "Scientific notation (e)",
			format:   "Sci: %e",
			args:     []any{1234.0},
			expected: "Sci: 1.234000e+03",
		},
		{
			name:     "Scientific notation (E)",
			format:   "Sci: %E",
			args:     []any{1234.0},
			expected: "Sci: 1.234000E+03",
		},
		{
			name:     "Compact float (g)",
			format:   "Compact: %g",
			args:     []any{1234.0},
			expected: "Compact: 1234",
		},
		{
			name:     "Compact float (G)",
			format:   "Compact: %G",
			args:     []any{1234.0},
			expected: "Compact: 1234",
		},
		{
			name:     "Pointer formatting",
			format:   "Pointer: %p",
			args:     []any{new(int)},
			expected: "Pointer: 0x",
		},
		{
			name:     "Hexadecimal uppercase",
			format:   "Hex: %X",
			args:     []any{255},
			expected: "Hex: FF",
		},
		{
			name:     "Octal uppercase",
			format:   "Octal: %O",
			args:     []any{64},
			expected: "Octal: 100",
		},
		{
			name:     "Binary uppercase",
			format:   "Binary: %B",
			args:     []any{7},
			expected: "Binary: 111",
		},
		{
			name:     "Unicode format",
			format:   "Unicode: %U",
			args:     []any{'A'},
			expected: "Unicode: U+0041",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Sprintf(test.format, test.args...)
			if out != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, out)
			}
		})
	}
}

func TestReporterFormatting(t *testing.T) {
	// ...existing tests...
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "Peak reduction percentage",
			format:   "- 🏆 **Peak Reduction: %.1f%%** (Best optimization)\n",
			args:     []any{71.5},
			expected: "- 🏆 **Peak Reduction: 71.5%** (Best optimization)\n",
		},
		{
			name:     "Average WebAssembly reduction",
			format:   "- ✅ **Average WebAssembly Reduction: %.1f%%**\n",
			args:     []any{53.2},
			expected: "- ✅ **Average WebAssembly Reduction: 53.2%**\n",
		},
		{
			name:     "Size savings with string",
			format:   "- 📦 **Total Size Savings: %s across all builds**\n\n",
			args:     []any{"1.7 MB"},
			expected: "- 📦 **Total Size Savings: 1.7 MB across all builds**\n\n",
		},
		{
			name:     "Memory efficiency class",
			format:   "- 💾 **Memory Efficiency**: %s (%.1f%% average change)\n",
			args:     []any{"❌ **Poor** (Significant overhead)", 154.2},
			expected: "- 💾 **Memory Efficiency**: ❌ **Poor** (Significant overhead) (154.2% average change)\n",
		},
		{
			name:     "Allocation efficiency class",
			format:   "- 🔢 **Allocation Efficiency**: %s (%.1f%% average change)\n",
			args:     []any{"❌ **Poor** (Excessive allocations)", 118.4},
			expected: "- 🔢 **Allocation Efficiency**: ❌ **Poor** (Excessive allocations) (118.4% average change)\n",
		},
		{
			name:     "Benchmarks analyzed count",
			format:   "- 📊 **Benchmarks Analyzed**: %d categories\n",
			args:     []any{3},
			expected: "- 📊 **Benchmarks Analyzed**: 3 categories\n",
		},
		{
			name:     "Complex table row with multiple formats",
			format:   "| %s **%s** | 📊 Standard | `%s` | `%d` | `%s` | - | - | - |\n",
			args:     []any{"📝", "String Processing", "1.2 KB", 48, "3.4μs"},
			expected: "| 📝 **String Processing** | 📊 Standard | `1.2 KB` | `48` | `3.4μs` | - | - | - |\n",
		},
		{
			name:     "fmt performance row",
			format:   "| | 🚀 fmt | `%s` | `%d` | `%s` | %s **%s** | %s **%s** | %s |\n",
			args:     []any{"2.8 KB", 119, "13.7μs", "❌", "140.3% more", "❌", "147.9% more", "❌ **Poor**"},
			expected: "| | 🚀 fmt | `2.8 KB` | `119` | `13.7μs` | ❌ **140.3% more** | ❌ **147.9% more** | ❌ **Poor** |\n",
		},
		{
			name:     "Binary size table row",
			format:   "| %s **%s Native** | `%s` | %s | %s | **-%s** | %s **%.1f%%** |\n",
			args:     []any{"🖥️", "Default", "-ldflags=\"-s -w\"", "1.3 MB", "1.1 MB", "176.0 KB", "➖", 13.4},
			expected: "| 🖥️ **Default Native** | `-ldflags=\"-s -w\"` | 1.3 MB | 1.1 MB | **-176.0 KB** | ➖ **13.4%** |\n",
		},
		{
			name:     "Error message formatting",
			format:   "Failed to read README: %v",
			args:     []any{Err("file not found")},
			expected: "Failed to read README: file not found",
		},
		{
			name:     "Memory improvement percentage",
			format:   "%.1f%% less",
			args:     []any{44.2},
			expected: "44.2% less",
		},
		{
			name:     "Memory improvement percentage more",
			format:   "%.1f%% more",
			args:     []any{140.3},
			expected: "140.3% more",
		},
		{
			name:     "Nanosecond formatting",
			format:   "%dns",
			args:     []any{int64(500)},
			expected: "500ns",
		},
		{
			name:     "Microsecond formatting",
			format:   "%.1fμs",
			args:     []any{3.4},
			expected: "3.4μs",
		},
		{
			name:     "Millisecond formatting",
			format:   "%.1fms",
			args:     []any{1.5},
			expected: "1.5ms",
		},
		{
			name:     "Alineación y ancho de campo: string",
			format:   "%-20s %-8s %-12s %-10s",
			args:     []any{"File", "Type", "Library", "Size"},
			expected: "File                 Type     Library      Size      ",
		},
		{
			name:     "Alineación y ancho de campo: valores",
			format:   "%-20s %-8s %-12s %-10s",
			args:     []any{"main.go", "native", "tinystring", "1.2MB"},
			expected: "main.go              native   tinystring   1.2MB     ",
		},
		{
			name:     "Alineación y ancho de campo: numérico",
			format:   "%8d %8d",
			args:     []any{123, 4567},
			expected: "     123     4567",
		},
		{
			name:     "Zero-padded integer",
			format:   "%02d",
			args:     []any{1},
			expected: "01",
		},
		{
			name:     "Zero-padded integer with larger width",
			format:   "%03d",
			args:     []any{5},
			expected: "005",
		},
		{
			name:     "Zero-padded integer already at width",
			format:   "%02d",
			args:     []any{12},
			expected: "12",
		},
		{
			name:     "Zero-padded integer with multiple digits",
			format:   "%05d",
			args:     []any{123},
			expected: "00123",
		},
		{
			name:     "Alineación y ancho de campo: mixto",
			format:   "%-10s %8d",
			args:     []any{"Total:", 99},
			expected: "Total:           99",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Sprintf(test.format, test.args...)
			if out != test.expected {
				t.Errorf("\nExpected:\n%q\ngot:\n%q", test.expected, out)
			}
		})
	}

	// Casos adicionales para formatos de logging y tamaño de bytes usados en common.go
	t.Run("LogStep format", func(t *testing.T) {
		msg := "Processing file"
		out := Sprintf("📋 %s\n", msg)
		expected := "📋 Processing file\n"
		if out != expected {
			t.Errorf("Expected %q, got %q", expected, out)
		}
	})
	t.Run("LogSuccess format", func(t *testing.T) {
		msg := "Completed successfully"
		out := Sprintf("✅ %s\n", msg)
		expected := "✅ Completed successfully\n"
		if out != expected {
			t.Errorf("Expected %q, got %q", expected, out)
		}
	})
	t.Run("LogError format", func(t *testing.T) {
		msg := "Something went wrong"
		out := Sprintf("❌ %s\n", msg)
		expected := "❌ Something went wrong\n"
		if out != expected {
			t.Errorf("Expected %q, got %q", expected, out)
		}
	})
	t.Run("LogInfo format", func(t *testing.T) {
		msg := "This is info"
		out := Sprintf("ℹ️  %s\n", msg)
		expected := "ℹ️  This is info\n"
		if out != expected {
			t.Errorf("Expected %q, got %q", expected, out)
		}
	})
	t.Run("FormatSize bytes", func(t *testing.T) {
		out := Sprintf("%d B", 512)
		expected := "512 B"
		if out != expected {
			t.Errorf("Expected %q, got %q", expected, out)
		}
	})
	t.Run("FormatSize kilobytes", func(t *testing.T) {
		out := Sprintf("%.1f %cB", 1.5, 'K')
		expected := "1.5 KB"
		if out != expected {
			t.Errorf("Expected %q, got %q", expected, out)
		}
	})
}

func TestFormatLocStr(t *testing.T) {
	// Define a test LocStr
	testLoc := LocStr{
		EN: "Hello",
		ES: "Hola",
	}

	// Save original language to restore later
	origLang := OutLang()
	defer OutLang(origLang)

	// Test default language (EN)
	OutLang(EN)
	if got := Sprintf("%L", testLoc); got != "Hello" {
		t.Errorf("Sprintf(%%L) EN = %q, want %q", got, "Hello")
	}

	// Test Spanish
	OutLang(ES)
	if got := Sprintf("%L", testLoc); got != "Hola" {
		t.Errorf("Sprintf(%%L) ES = %q, want %q", got, "Hola")
	}

	// Test pointer to LocStr
	OutLang(EN)
	if got := Sprintf("%L", &testLoc); got != "Hello" {
		t.Errorf("Sprintf(%%L) *LocStr = %q, want %q", got, "Hello")
	}

	// Test mixed
	OutLang(EN)
	if got := Sprintf("Say %L world", testLoc); got != "Say Hello world" {
		t.Errorf("Fmt mixed = %q, want %q", got, "Say Hello world")
	}
}
