package response

import (
	"bytes"
	"strings"
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		env  Envelope
		want int
	}{
		{name: "ok", env: OK(map[string]any{"status": "ok"}, nil), want: 0},
		{name: "usage", env: Fail("usage_error", "bad input", nil, nil), want: 2},
		{name: "dependency", env: Fail("dependency_error", "missing dep", nil, nil), want: 3},
		{name: "not_found", env: Fail("not_found", "missing", nil, nil), want: 4},
		{name: "lock_conflict", env: Fail("lock_conflict", "locked", nil, nil), want: 5},
		{name: "timeout", env: Fail("timeout", "slow", nil, nil), want: 6},
		{name: "upstream", env: Fail("upstream_error", "upstream failed", nil, nil), want: 7},
		{name: "unknown", env: Fail("random_error", "unknown", nil, nil), want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCode(tt.env); got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

type fakeTextRenderer struct{}

func (fakeTextRenderer) RenderText() string {
	return "rendered text output"
}

func TestPrintTextUsesRenderer(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := printText(&stdout, &stderr, OK(fakeTextRenderer{}, nil)); err != nil {
		t.Fatalf("printText() error = %v", err)
	}
	if got := stdout.String(); got != "rendered text output\n" {
		t.Fatalf("stdout = %q, want rendered output", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPrintTextErrorIncludesMessage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := printText(&stdout, &stderr, Fail("not_found", "missing session", nil, nil)); err != nil {
		t.Fatalf("printText() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, "not_found: missing session") {
		t.Fatalf("stderr = %q, want error message", got)
	}
}
