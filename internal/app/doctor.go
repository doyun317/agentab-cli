package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/agentab/agentab-cli/internal/state"
)

type doctorReport struct {
	AgentabHome      string              `json:"agentabHome"`
	ArtifactsDir     string              `json:"artifactsDir"`
	ManagedBinPath   string              `json:"managedBinPath"`
	PinchtabURL      string              `json:"pinchtabURL"`
	PinchtabHealthy  bool                `json:"pinchtabHealthy"`
	PinchtabBin      string              `json:"pinchtabBin,omitempty"`
	PinchtabManaged  *bool               `json:"pinchtabManaged,omitempty"`
	PinchtabBinError string              `json:"pinchtabBinError,omitempty"`
	ChromeBin        string              `json:"chromeBin"`
	ChromeBinFound   bool                `json:"chromeBinFound"`
	ChromeBinSource  string              `json:"chromeBinSource"`
	ChromeBinError   string              `json:"chromeBinError,omitempty"`
	Daemon           *state.DaemonInfo   `json:"daemon,omitempty"`
	Pinchtab         *state.PinchtabInfo `json:"pinchtab,omitempty"`
}

func (r doctorReport) RenderText() string {
	var b strings.Builder

	b.WriteString("agentab doctor\n")
	b.WriteString(fmt.Sprintf("home: %s\n", valueOrFallback(r.AgentabHome, "-")))
	b.WriteString(fmt.Sprintf("artifacts: %s\n", valueOrFallback(r.ArtifactsDir, "-")))
	b.WriteString(fmt.Sprintf("managed pinchtab bin: %s\n", valueOrFallback(r.ManagedBinPath, "-")))

	b.WriteString("\nchrome\n")
	b.WriteString(fmt.Sprintf("  status: %s\n", statusWord(r.ChromeBinFound)))
	b.WriteString(fmt.Sprintf("  source: %s\n", valueOrFallback(r.ChromeBinSource, "-")))
	b.WriteString(fmt.Sprintf("  path: %s\n", valueOrFallback(r.ChromeBin, "-")))
	if r.ChromeBinError != "" {
		b.WriteString(fmt.Sprintf("  error: %s\n", r.ChromeBinError))
	}

	b.WriteString("\npinchtab\n")
	b.WriteString(fmt.Sprintf("  health: %s\n", statusWord(r.PinchtabHealthy)))
	b.WriteString(fmt.Sprintf("  url: %s\n", valueOrFallback(r.PinchtabURL, "-")))
	if r.PinchtabBin != "" {
		b.WriteString(fmt.Sprintf("  binary: %s\n", r.PinchtabBin))
	}
	if r.PinchtabManaged != nil {
		b.WriteString(fmt.Sprintf("  managed binary: %s\n", yesNo(*r.PinchtabManaged)))
	}
	if r.PinchtabBinError != "" {
		b.WriteString(fmt.Sprintf("  binary error: %s\n", r.PinchtabBinError))
	}
	if r.Pinchtab != nil {
		b.WriteString(fmt.Sprintf("  runtime pid: %s\n", intValueOrFallback(r.Pinchtab.PID)))
		if !r.Pinchtab.StartedAt.IsZero() {
			b.WriteString(fmt.Sprintf("  runtime started: %s\n", formatTime(r.Pinchtab.StartedAt)))
		}
		b.WriteString(fmt.Sprintf("  runtime token: %s\n", tokenState(r.Pinchtab.Token)))
	}

	b.WriteString("\ndaemon\n")
	if r.Daemon == nil {
		b.WriteString("  status: not running\n")
	} else {
		b.WriteString("  status: running\n")
		b.WriteString(fmt.Sprintf("  pid: %s\n", intValueOrFallback(r.Daemon.PID)))
		b.WriteString(fmt.Sprintf("  port: %s\n", intValueOrFallback(r.Daemon.Port)))
		if !r.Daemon.StartedAt.IsZero() {
			b.WriteString(fmt.Sprintf("  started: %s\n", formatTime(r.Daemon.StartedAt)))
		}
		b.WriteString(fmt.Sprintf("  token: %s\n", tokenState(r.Daemon.Token)))
	}

	return strings.TrimRight(b.String(), "\n")
}

func boolPtr(v bool) *bool {
	return &v
}

func statusWord(ok bool) string {
	if ok {
		return "ok"
	}
	return "not ready"
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func valueOrFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func intValueOrFallback(value int) string {
	if value == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", value)
}

func tokenState(token string) string {
	if strings.TrimSpace(token) == "" {
		return "missing"
	}
	return "present"
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.UTC().Format(time.RFC3339)
}
