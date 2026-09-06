package fmt_test

import (
	"testing"
	tf "webtyp.com/fmt"
)

func TestCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		camelUp  string
		camelLow string
	}{
		{"id", "Id", "id"},
		{"sku", "Sku", "sku"},
		{"tenant_id", "TenantId", "tenantId"},
		{"is_active", "IsActive", "isActive"},
		{"updated_at", "UpdatedAt", "updatedAt"},
		{"api_response", "ApiResponse", "apiResponse"},
		{"user123_name", "User123Name", "user123Name"},
		{"name", "Name", "name"},
		{"first-name", "FirstName", "firstName"},
		{"hello world", "HelloWorld", "helloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			up := tf.Convert(tt.input).CamelUp().String()
			if up != tt.camelUp {
				t.Errorf("CamelUp(%q) = %q, want %q", tt.input, up, tt.camelUp)
			}

			low := tf.Convert(tt.input).CamelLow().String()
			if low != tt.camelLow {
				t.Errorf("CamelLow(%q) = %q, want %q", tt.input, low, tt.camelLow)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		"id",
		"sku",
		"tenant_id",
		"is_active",
		"updated_at",
		"api_response",
		"user123_name",
		"name",
		"first_name",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			got := tf.Convert(tt).CamelUp().SnakeLow().String()
			if got != tt {
				t.Errorf("SnakeLow(CamelUp(%q)) = %q, want %q", tt, got, tt)
			}
		})
	}
}

func TestAcronymSnake(t *testing.T) {
    // APIResponse -> api_response
    input := "APIResponse"
    want := "api_response"
    got := tf.Convert(input).SnakeLow().String()
    if got != want {
        t.Errorf("SnakeLow(%q) = %q, want %q", input, got, want)
    }

    // HTTPServer -> http_server
    input = "HTTPServer"
    want = "http_server"
    got = tf.Convert(input).SnakeLow().String()
    if got != want {
        t.Errorf("SnakeLow(%q) = %q, want %q", input, got, want)
    }
}

func TestSeparatorUnification(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"first-name", "first_name"},
		{"first_name", "first-name"}, // using SnakeLow("-")
		{"UserID", "user_id"},
		{"user_id", "UserId"}, // using CamelUp()
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var got string
			if tt.want == "first-name" {
				got = tf.Convert(tt.input).SnakeLow("-").String()
			} else if tt.want == "UserId" {
				got = tf.Convert(tt.input).CamelUp().String()
			} else {
				got = tf.Convert(tt.input).SnakeLow().String()
			}

			if got != tt.want {
				t.Errorf("Convert(%q) to %q failed, got %q", tt.input, tt.want, got)
			}
		})
	}
}
