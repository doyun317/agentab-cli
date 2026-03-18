package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type Session struct {
	Name         string    `json:"name"`
	InstanceID   string    `json:"instanceId"`
	ProfileID    string    `json:"profileId,omitempty"`
	Mode         string    `json:"mode,omitempty"`
	LastUsedAt   time.Time `json:"lastUsedAt"`
	CurrentTabID string    `json:"currentTabId,omitempty"`
}

type PersistedState struct {
	CurrentSession string             `json:"currentSession,omitempty"`
	Sessions       map[string]Session `json:"sessions"`
}

type DaemonInfo struct {
	Port      int       `json:"port"`
	Token     string    `json:"token"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"startedAt"`
}

type PinchtabInfo struct {
	BaseURL   string    `json:"baseURL"`
	Token     string    `json:"token,omitempty"`
	PID       int       `json:"pid,omitempty"`
	StartedAt time.Time `json:"startedAt,omitempty"`
}

type Store struct {
	root         string
	statePath    string
	daemonPath   string
	pinchtabPath string
	logsDir      string
	artifactsDir string
	binDir       string
	runDir       string
	mu           sync.Mutex
}

func HomeDir() (string, error) {
	if root := os.Getenv("AGENTAB_HOME"); root != "" {
		return root, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("discover home dir: %w", err)
	}
	return filepath.Join(home, ".agentab"), nil
}

func NewStore(root string) (*Store, error) {
	var err error
	if root == "" {
		root, err = HomeDir()
		if err != nil {
			return nil, err
		}
	}
	store := &Store{
		root:         root,
		statePath:    filepath.Join(root, "state.json"),
		daemonPath:   filepath.Join(root, "run", "daemon.json"),
		pinchtabPath: filepath.Join(root, "run", "pinchtab.json"),
		logsDir:      filepath.Join(root, "logs"),
		artifactsDir: filepath.Join(root, "artifacts"),
		binDir:       filepath.Join(root, "bin"),
		runDir:       filepath.Join(root, "run"),
	}
	if err := store.ensureLayout(); err != nil {
		return nil, err
	}
	if _, err := os.Stat(store.statePath); errors.Is(err, os.ErrNotExist) {
		if err := store.SaveState(PersistedState{Sessions: map[string]Session{}}); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *Store) Root() string         { return s.root }
func (s *Store) LogsDir() string      { return s.logsDir }
func (s *Store) ArtifactsDir() string { return s.artifactsDir }
func (s *Store) BinDir() string       { return s.binDir }
func (s *Store) RunDir() string       { return s.runDir }

func (s *Store) ensureLayout() error {
	for _, dir := range []string{s.root, s.logsDir, s.artifactsDir, s.binDir, s.runDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}

func (s *Store) LoadState() (PersistedState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked()
}

func (s *Store) SaveState(st PersistedState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st.Sessions == nil {
		st.Sessions = map[string]Session{}
	}
	return writeJSONAtomic(s.statePath, st)
}

func (s *Store) ListSessions() ([]Session, error) {
	st, err := s.LoadState()
	if err != nil {
		return nil, err
	}
	items := make([]Session, 0, len(st.Sessions))
	for _, session := range st.Sessions {
		items = append(items, session)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items, nil
}

func (s *Store) GetSession(name string) (Session, error) {
	st, err := s.LoadState()
	if err != nil {
		return Session{}, err
	}
	session, ok := st.Sessions[name]
	if !ok {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (s *Store) CurrentSession() (Session, error) {
	st, err := s.LoadState()
	if err != nil {
		return Session{}, err
	}
	if st.CurrentSession == "" {
		return Session{}, ErrNotFound
	}
	session, ok := st.Sessions[st.CurrentSession]
	if !ok {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (s *Store) PutSession(session Session) error {
	st, err := s.LoadState()
	if err != nil {
		return err
	}
	if st.Sessions == nil {
		st.Sessions = map[string]Session{}
	}
	st.Sessions[session.Name] = session
	st.CurrentSession = session.Name
	return s.SaveState(st)
}

func (s *Store) DeleteSession(name string) error {
	st, err := s.LoadState()
	if err != nil {
		return err
	}
	delete(st.Sessions, name)
	if st.CurrentSession == name {
		st.CurrentSession = ""
	}
	return s.SaveState(st)
}

func (s *Store) SetCurrentSession(name string) error {
	st, err := s.LoadState()
	if err != nil {
		return err
	}
	if _, ok := st.Sessions[name]; !ok {
		return ErrNotFound
	}
	st.CurrentSession = name
	return s.SaveState(st)
}

func (s *Store) UpdateSession(name string, update func(*Session) error) (Session, error) {
	st, err := s.LoadState()
	if err != nil {
		return Session{}, err
	}
	session, ok := st.Sessions[name]
	if !ok {
		return Session{}, ErrNotFound
	}
	if err := update(&session); err != nil {
		return Session{}, err
	}
	st.Sessions[name] = session
	if err := s.SaveState(st); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *Store) ClearSessions() error {
	return s.SaveState(PersistedState{
		CurrentSession: "",
		Sessions:       map[string]Session{},
	})
}

func (s *Store) ReadDaemonInfo() (DaemonInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var info DaemonInfo
	raw, err := os.ReadFile(s.daemonPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DaemonInfo{}, ErrNotFound
		}
		return DaemonInfo{}, err
	}
	if err := json.Unmarshal(raw, &info); err != nil {
		return DaemonInfo{}, err
	}
	return info, nil
}

func (s *Store) WriteDaemonInfo(info DaemonInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeJSONAtomic(s.daemonPath, info)
}

func (s *Store) ClearDaemonInfo() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.daemonPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) ReadPinchtabInfo() (PinchtabInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var info PinchtabInfo
	raw, err := os.ReadFile(s.pinchtabPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PinchtabInfo{}, ErrNotFound
		}
		return PinchtabInfo{}, err
	}
	if err := json.Unmarshal(raw, &info); err != nil {
		return PinchtabInfo{}, err
	}
	return info, nil
}

func (s *Store) WritePinchtabInfo(info PinchtabInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeJSONAtomic(s.pinchtabPath, info)
}

func (s *Store) ClearPinchtabInfo() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.pinchtabPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) loadUnlocked() (PersistedState, error) {
	var st PersistedState
	raw, err := os.ReadFile(s.statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PersistedState{Sessions: map[string]Session{}}, nil
		}
		return PersistedState{}, err
	}
	if err := json.Unmarshal(raw, &st); err != nil {
		return PersistedState{}, err
	}
	if st.Sessions == nil {
		st.Sessions = map[string]Session{}
	}
	return st, nil
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
