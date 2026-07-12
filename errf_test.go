package fmt

import "testing"

// Errf used to truncate any message that had literal text before a verb: the formatted
// value and everything after it vanished, with no error reported.
//
// Cause: wrFormat writes into dest, and for Errf dest IS BuffErr. After formatting a value
// it checked `hasContent(BuffErr)` to detect a formatting failure — but the literal prefix
// already sitting in BuffErr made that check a false positive, so it bailed out early.
//
// The failure was SILENT: no panic, no error, just a diagnostic message with its subject
// missing. `Errf("filetype: %s is not allowed", "PDF")` returned "filetype: ".
func TestErrfKeepsTextAroundVerbs(t *testing.T) {
	tests := []struct {
		name   string
		got    string
		expect string
	}{
		{
			name:   "literal text before the verb",
			got:    Errf("filetype: %s is not allowed", "PDF").Error(),
			expect: "filetype: PDF is not allowed",
		},
		{
			name:   "verb first (this case always worked: BuffErr was still empty)",
			got:    Errf("%s is not allowed", "PDF").Error(),
			expect: "PDF is not allowed",
		},
		{
			name:   "no verbs at all",
			got:    Errf("plain text, no verbs").Error(),
			expect: "plain text, no verbs",
		},
		{
			name:   "several verbs with text between them",
			got:    Errf("db: row %d of table %s failed", 7, "users").Error(),
			expect: "db: row 7 of table users failed",
		},
		{
			name:   "trailing text after the last verb",
			got:    Errf("port %d is busy", 8080).Error(),
			expect: "port 8080 is busy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expect {
				t.Errorf("Errf() = %q, want %q", tt.got, tt.expect)
			}
		})
	}
}

// An unsupported verb must still report loudly — fixing the truncation must not turn a
// noisy failure into a silent one.
func TestErrfUnsupportedVerbStillReports(t *testing.T) {
	got := Errf("scan: %T", 1).Error()
	if got == "scan: " || got == "" {
		t.Fatalf("Errf() = %q — an unsupported verb must not fail silently", got)
	}
}
