package cmd

import "testing"

func TestFormatFrontmatterValueQuotesUnsafeScalars(t *testing.T) {
	if got := formatFrontmatterValue("plain"); got != "plain" {
		t.Fatalf("plain scalar should stay unquoted, got %q", got)
	}
	if got := formatFrontmatterValue("path: 12"); got != `"path: 12"` {
		t.Fatalf("colon-space scalar should be quoted, got %q", got)
	}
	if got := formatFrontmatterValue("line1\nline2"); got != `"line1\nline2"` {
		t.Fatalf("newline scalar should be quoted, got %q", got)
	}
	if got := formatFrontmatterValue("---"); got != `"---"` {
		t.Fatalf("document marker should be quoted, got %q", got)
	}
}
