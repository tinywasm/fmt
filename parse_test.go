package fmt_test

import (
	"testing"

	. "github.com/tinywasm/fmt"
)

func TestParseKeyValue(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		delimiter   string
		wantValue   string
		wantErrText string
	}{
		{
			name:        "Basic key-value with default delimiter",
			input:       "name:John",
			delimiter:   "",
			wantValue:   "John",
			wantErrText: "",
		},
		{
			name:        "No delimiter in string",
			input:       "invalid-string",
			delimiter:   "",
			wantValue:   "",
			wantErrText: Translate(D.Format, D.Invalid, D.Delimiter, D.Not, D.Found).String(),
		},
		{
			name:        "Custom delimiter",
			input:       "age=30",
			delimiter:   "=",
			wantValue:   "30",
			wantErrText: "",
		},
		{
			name:        "Value contains delimiter",
			input:       "address:123 Main St:Apt 4",
			delimiter:   "",
			wantValue:   "123 Main St:Apt 4",
			wantErrText: "",
		},
		{
			name:        "Empty input",
			input:       "",
			delimiter:   "",
			wantValue:   "",
			wantErrText: Translate(D.Format, D.Invalid, D.Delimiter, D.Not, D.Found).String(),
		},
		{
			name:        "Only delimiter",
			input:       ":",
			delimiter:   "",
			wantValue:   "",
			wantErrText: "",
		},
		{
			name:        "Multi-character delimiter",
			input:       "key=>value",
			delimiter:   "=>",
			wantValue:   "value",
			wantErrText: "",
		},
		{
			name:        "Missing custom delimiter",
			input:       "key:value",
			delimiter:   "=",
			wantValue:   "",
			wantErrText: Translate(D.Format, D.Invalid, D.Delimiter, D.Not, D.Found).String(),
		},
		{
			name:        "Empty delimiter uses default",
			input:       "name:John",
			delimiter:   "",
			wantValue:   "John",
			wantErrText: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var delimiters []string
			if tc.delimiter != "" {
				delimiters = append(delimiters, tc.delimiter)
			}

			gotValue, gotErr := Convert(tc.input).ExtractValue(delimiters...)

			if gotValue != tc.wantValue {
				t.Errorf("ParseKeyValue() value = %q, want %q", gotValue, tc.wantValue)
			}

			// Check error
			if tc.wantErrText == "" {
				if gotErr != nil {
					t.Errorf("ParseKeyValue() error = %v, want nil", gotErr)
				}
			} else {
				if gotErr == nil {
					t.Errorf("ParseKeyValue() error = nil, want error containing %q", tc.wantErrText)
				} else if !Contains(gotErr.Error(), tc.wantErrText) {
					t.Errorf("ParseKeyValue() error = %v, want error containing %q", gotErr, tc.wantErrText)
				}
			}
		})
	}
}

func TestTagValue(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name:      "Basic tag value extraction",
			input:     `json:"name"`,
			key:       "json",
			wantValue: "name",
			wantFound: true,
		},
		{
			name:      "Multiple tags with target in middle",
			input:     `json:"name" Label:"Nombre" xml:"nm"`,
			key:       "Label",
			wantValue: "Nombre",
			wantFound: true,
		},
		{
			name:      "Multiple tags with target at end",
			input:     `json:"name" Label:"Nombre" xml:"nm"`,
			key:       "xml",
			wantValue: "nm",
			wantFound: true,
		},
		{
			name:      "Multiple tags with target at start",
			input:     `json:"name" Label:"Nombre" xml:"nm"`,
			key:       "json",
			wantValue: "name",
			wantFound: true,
		},
		{
			name:      "Key not found",
			input:     `json:"name" Label:"Nombre"`,
			key:       "xml",
			wantValue: "",
			wantFound: false,
		},
		{
			name:      "Empty input",
			input:     "",
			key:       "json",
			wantValue: "",
			wantFound: false,
		},
		{
			name:      "No quotes in value",
			input:     `json:name`,
			key:       "json",
			wantValue: "name",
			wantFound: true,
		},
		{
			name:      "Extra spaces between tags",
			input:     `json:"name"   Label:"Nombre"    xml:"nm"`,
			key:       "Label",
			wantValue: "Nombre",
			wantFound: true,
		},
		{
			name:      "Tag without colon",
			input:     `json:"name" invalid Label:"Nombre"`,
			key:       "Label",
			wantValue: "Nombre",
			wantFound: true,
		},
		{
			name:      "Complex struct tag",
			input:     `json:"user_name,omitempty" validate:"required,min=3" db:"username"`,
			key:       "validate",
			wantValue: "required,min=3",
			wantFound: true,
		},
		{
			name:      "Single tag",
			input:     `json:"name"`,
			key:       "json",
			wantValue: "name",
			wantFound: true,
		},
		{
			name:      "Empty quotes",
			input:     `json:""`,
			key:       "json",
			wantValue: "",
			wantFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotValue, gotFound := Convert(tc.input).TagValue(tc.key)

			if gotValue != tc.wantValue {
				t.Errorf("TagValue() value = %q, want %q", gotValue, tc.wantValue)
			}

			if gotFound != tc.wantFound {
				t.Errorf("TagValue() found = %v, want %v", gotFound, tc.wantFound)
			}
		})
	}
}
func TestTagPairs(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		key       string
		wantPairs []KeyValue
	}{
		{
			name:      "Basic options extraction",
			input:     `options:"key1:text1,key2:text2"`,
			key:       "options",
			wantPairs: []KeyValue{{Key: "key1", Value: "text1"}, {Key: "key2", Value: "text2"}},
		},
		{
			name:      "Options with other tags",
			input:     `json:"name" options:"key1:text1,key2:text2" validate:"required"`,
			key:       "options",
			wantPairs: []KeyValue{{Key: "key1", Value: "text1"}, {Key: "key2", Value: "text2"}},
		},
		{
			name:      "Single option",
			input:     `options:"key1:text1"`,
			key:       "options",
			wantPairs: []KeyValue{{Key: "key1", Value: "text1"}},
		},
		{
			name:      "Empty options tag",
			input:     `options:""`,
			key:       "options",
			wantPairs: nil,
		},
		{
			name:      "Key not found",
			input:     `json:"name"`,
			key:       "options",
			wantPairs: nil,
		},
		{
			name:      "Malfomed pairs (missing colon)",
			input:     `options:"key1,key2:text2"`,
			key:       "options",
			wantPairs: []KeyValue{{Key: "key2", Value: "text2"}},
		},
		{
			name:      "Multiple pairs without spaces (standard format)",
			input:     `options:"key1:text1,key2:text2"`,
			key:       "options",
			wantPairs: []KeyValue{{Key: "key1", Value: "text1"}, {Key: "key2", Value: "text2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPairs := Convert(tc.input).TagPairs(tc.key)

			if len(gotPairs) != len(tc.wantPairs) {
				t.Fatalf("TagPairs() length = %d, want %d", len(gotPairs), len(tc.wantPairs))
			}

			for i := range gotPairs {
				if gotPairs[i].Key != tc.wantPairs[i].Key || gotPairs[i].Value != tc.wantPairs[i].Value {
					t.Errorf("TagPairs()[%d] = %v, want %v", i, gotPairs[i], tc.wantPairs[i])
				}
			}
		})
	}
}
