package fmt

import "testing"

func TestIsZero(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"Nil", nil, true},
		{"StringEmpty", "", true},
		{"StringNotEmpty", "x", false},
		{"BoolFalse", false, true},
		{"BoolTrue", true, false},
		{"IntZero", 0, true},
		{"IntNotZero", 1, false},
		{"Int8Zero", int8(0), true},
		{"Int16Zero", int16(0), true},
		{"Int32Zero", int32(0), true},
		{"Int64Zero", int64(0), true},
		{"UintZero", uint(0), true},
		{"Uint8Zero", uint8(0), true},
		{"Uint16Zero", uint16(0), true},
		{"Uint32Zero", uint32(0), true},
		{"Uint64Zero", uint64(0), true},
		{"Float32Zero", float32(0), true},
		{"Float32NotZero", float32(1.5), false},
		{"Float64Zero", float64(0), true},
		{"Float64NotZero", float64(1.5), false},
		{"BytesEmpty", []byte{}, true},
		{"BytesNotEmpty", []byte{1}, false},
		{"UnknownType", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZero(tt.input); got != tt.expected {
				t.Errorf("IsZero(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
