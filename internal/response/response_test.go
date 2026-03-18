package response

import "testing"

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
