package netmount

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spalencsar/omarchy-files/core/internal/atomicfile"
)

const recentServers = 12

type Server struct {
	URI      string `json:"uri"`
	Name     string `json:"name,omitempty"`
	Scheme   string `json:"scheme,omitempty"`
	LastUsed int64  `json:"lastUsedMs"`
}

type serverStore struct {
	path string
	now  func() time.Time

	mu     sync.Mutex
	loaded bool
	items  []Server
}

func newServerStore(configDir string) *serverStore {
	return &serverStore{path: filepath.Join(configDir, "servers.json"), now: time.Now}
}

func (s *serverStore) list() []Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load()
	return append([]Server(nil), s.items...)
}

func (s *serverStore) remember(server Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load()

	server.LastUsed = s.now().UnixMilli()
	kept := []Server{server}
	for _, existing := range s.items {
		if existing.URI == server.URI || len(kept) >= recentServers {
			continue
		}
		kept = append(kept, existing)
	}
	s.items = kept
	return s.save()
}

func (s *serverStore) forget(uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load()

	kept := make([]Server, 0, len(s.items))
	for _, existing := range s.items {
		if existing.URI == uri {
			continue
		}
		kept = append(kept, existing)
	}
	s.items = kept
	return s.save()
}

func (s *serverStore) load() {
	if s.loaded {
		return
	}
	s.loaded = true

	raw, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, &s.items)
}

func (s *serverStore) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(s.path, raw, 0o600)
}
