package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionPersistence(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	session := Session{
		Name:         "demo",
		InstanceID:   "inst_123",
		ProfileID:    "prof_123",
		Mode:         "headless",
		LastUsedAt:   time.Unix(1700000000, 0).UTC(),
		CurrentTabID: "tab_1",
	}
	if err := store.PutSession(session); err != nil {
		t.Fatalf("PutSession() error = %v", err)
	}

	got, err := store.GetSession("demo")
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.InstanceID != session.InstanceID || got.CurrentTabID != session.CurrentTabID {
		t.Fatalf("GetSession() = %+v, want %+v", got, session)
	}

	current, err := store.CurrentSession()
	if err != nil {
		t.Fatalf("CurrentSession() error = %v", err)
	}
	if current.Name != "demo" {
		t.Fatalf("CurrentSession().Name = %q, want demo", current.Name)
	}
}

func TestDaemonInfoRoundTrip(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	info := DaemonInfo{
		Port:      43921,
		Token:     "secret",
		PID:       4242,
		StartedAt: time.Unix(1700000000, 0).UTC(),
	}
	if err := store.WriteDaemonInfo(info); err != nil {
		t.Fatalf("WriteDaemonInfo() error = %v", err)
	}

	got, err := store.ReadDaemonInfo()
	if err != nil {
		t.Fatalf("ReadDaemonInfo() error = %v", err)
	}
	if got.Port != info.Port || got.Token != info.Token || got.PID != info.PID {
		t.Fatalf("ReadDaemonInfo() = %+v, want %+v", got, info)
	}

	if err := store.ClearDaemonInfo(); err != nil {
		t.Fatalf("ClearDaemonInfo() error = %v", err)
	}
	if _, err := store.ReadDaemonInfo(); err != ErrNotFound {
		t.Fatalf("ReadDaemonInfo() after clear error = %v, want %v", err, ErrNotFound)
	}
}

func TestPinchtabInfoRoundTrip(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	info := PinchtabInfo{
		BaseURL:   "http://127.0.0.1:9877",
		Token:     "secret",
		PID:       5151,
		StartedAt: time.Unix(1700000000, 0).UTC(),
	}
	if err := store.WritePinchtabInfo(info); err != nil {
		t.Fatalf("WritePinchtabInfo() error = %v", err)
	}

	got, err := store.ReadPinchtabInfo()
	if err != nil {
		t.Fatalf("ReadPinchtabInfo() error = %v", err)
	}
	if got.BaseURL != info.BaseURL || got.Token != info.Token || got.PID != info.PID {
		t.Fatalf("ReadPinchtabInfo() = %+v, want %+v", got, info)
	}

	if err := store.ClearPinchtabInfo(); err != nil {
		t.Fatalf("ClearPinchtabInfo() error = %v", err)
	}
	if _, err := store.ReadPinchtabInfo(); err != ErrNotFound {
		t.Fatalf("ReadPinchtabInfo() after clear error = %v, want %v", err, ErrNotFound)
	}
}

func TestNewStoreCreatesArtifactsDir(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	info, err := os.Stat(store.ArtifactsDir())
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", store.ArtifactsDir(), err)
	}
	if !info.IsDir() {
		t.Fatalf("ArtifactsDir() = %q, want directory", store.ArtifactsDir())
	}
	if filepath.Base(store.ArtifactsDir()) != "artifacts" {
		t.Fatalf("ArtifactsDir() = %q, want artifacts suffix", store.ArtifactsDir())
	}
	if got := filepath.Base(store.DaemonLogPath()); got != "agentab-daemon.log" {
		t.Fatalf("DaemonLogPath() = %q, want agentab-daemon.log suffix", got)
	}
	if got := filepath.Base(store.PinchtabLogPath()); got != "pinchtab.log" {
		t.Fatalf("PinchtabLogPath() = %q, want pinchtab.log suffix", got)
	}
}

func TestClearSessions(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if err := store.PutSession(Session{
		Name:         "demo",
		InstanceID:   "inst_123",
		CurrentTabID: "tab_1",
		LastUsedAt:   time.Unix(1700000000, 0).UTC(),
	}); err != nil {
		t.Fatalf("PutSession() error = %v", err)
	}

	if err := store.ClearSessions(); err != nil {
		t.Fatalf("ClearSessions() error = %v", err)
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("ListSessions() = %+v, want empty", sessions)
	}
	if _, err := store.CurrentSession(); err != ErrNotFound {
		t.Fatalf("CurrentSession() error = %v, want %v", err, ErrNotFound)
	}
}
