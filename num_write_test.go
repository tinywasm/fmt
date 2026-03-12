package fmt

import (
	"testing"
)

func TestWriteInt(t *testing.T) {
	tests := []struct {
		val  int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{-1, "-1"},
		{-123, "-123"},
		{9223372036854775807, "9223372036854775807"},
		{-9223372036854775808, "-9223372036854775808"},
	}

	for _, tt := range tests {
		c := Convert()
		c.WriteInt(tt.val)
		got := c.String()
		if got != tt.want {
			t.Errorf("WriteInt(%d) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestWriteFloat(t *testing.T) {
	tests := []struct {
		val  float64
		want string
	}{
		{0.0, "0"},
		{1.0, "1"},
		{1.1, "1.1"},
		{3.14, "3.14"},
		{-1.1, "-1.1"},
		{123.456, "123.456"},
	}

	for _, tt := range tests {
		c := Convert()
		c.WriteFloat(tt.val)
		got := c.String()
		if got != tt.want {
			t.Errorf("WriteFloat(%f) = %q, want %q", tt.val, got, tt.want)
		}
	}
}
