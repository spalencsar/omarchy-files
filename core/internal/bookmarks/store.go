package bookmarks

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/spalencsar/omarchy-files/core/internal/atomicfile"
	"github.com/duxlabs-dev/dux-go/log"
)

const (
	FileName       = "bookmarks.json"
	Topic          = "bookmarks"
	fileVersion    = 2
	coalesceWindow = 50 * time.Millisecond
)

type Publisher interface {
	Publish(topic string, data any)
}

type Bookmark struct {
	Path  string `json:"path"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

type changed struct {
	Kind      string     `json:"kind"`
	Bookmarks []Bookmark `json:"bookmarks"`
}

// Version 1 files were a bare array written before the sidebar places lived here.
type file struct {
	Version   int        `json:"version"`
	Bookmarks []Bookmark `json:"bookmarks"`
}

type Store struct {
	bus     Publisher
	dir     string
	file    string
	gtkFile string

	mu          sync.Mutex
	items       []Bookmark
	lastWritten []byte
	gtkImport   bool
	seeded      bool
	seedPaths   []string

	watcher *fsnotify.Watcher
	done    chan struct{}
	stopped sync.WaitGroup
}

func New(bus Publisher, configDir, gtkFile string) *Store {
	return &Store{bus: bus, dir: configDir, file: filepath.Join(configDir, FileName), gtkFile: gtkFile}
}

// The directory is watched, not the file: other editors replace it by rename.
func (s *Store) Start() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	s.mu.Lock()
	s.items, s.lastWritten, s.seeded, _ = s.load()
	s.gtkImport = s.gtkImportAvailable()
	s.mu.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(s.dir); err != nil {
		_ = watcher.Close()
		return err
	}
	s.watcher = watcher
	s.done = make(chan struct{})
	s.stopped.Add(1)
	go s.watch()
	return nil
}

func (s *Store) Close() {
	if s.watcher == nil {
		return
	}
	close(s.done)
	_ = s.watcher.Close()
	s.stopped.Wait()
	s.watcher = nil
}

func (s *Store) List() []Bookmark {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.items)
}

func (s *Store) Exists() bool {
	_, err := os.Stat(s.file)
	return err == nil
}

func (s *Store) Seeded() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seeded
}

// Seeding stays in memory so an unedited list keeps following the user dirs and an external legacy file is reseeded on reload.
func (s *Store) Seed(paths []string) ([]Bookmark, error) {
	for _, path := range paths {
		if !validPath(path) {
			return nil, errors.New("seed paths must be absolute")
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seedPaths = slices.Clone(paths)
	if s.seeded {
		return slices.Clone(s.items), nil
	}
	s.items = seed(s.seedPaths, s.items)
	s.seeded = true
	s.publish(s.items)
	return slices.Clone(s.items), nil
}

func seed(paths []string, items []Bookmark) []Bookmark {
	next := make([]Bookmark, 0, len(paths)+len(items))
	for _, path := range paths {
		path = clean(path)
		if indexOf(next, path) >= 0 {
			continue
		}
		if i := indexOf(items, path); i >= 0 {
			next = append(next, items[i])
			continue
		}
		next = append(next, Bookmark{Path: path, Label: labelOf(path)})
	}
	for _, b := range items {
		if indexOf(next, b.Path) < 0 {
			next = append(next, b)
		}
	}
	return next
}

func (s *Store) Add(b Bookmark) ([]Bookmark, error) {
	if !validPath(b.Path) {
		return nil, errors.New("bookmark path must be absolute")
	}
	b.Path = clean(b.Path)
	if b.Label == "" {
		b.Label = labelOf(b.Path)
	}
	return s.update(func(items []Bookmark) ([]Bookmark, error) {
		if indexOf(items, b.Path) >= 0 {
			return items, nil
		}
		return append(items, b), nil
	})
}

func (s *Store) Remove(path string) ([]Bookmark, error) {
	return s.update(func(items []Bookmark) ([]Bookmark, error) {
		i := indexOf(items, path)
		if i < 0 {
			return items, nil
		}
		return slices.Delete(items, i, i+1), nil
	})
}

func (s *Store) Rename(path, label string) ([]Bookmark, error) {
	return s.update(func(items []Bookmark) ([]Bookmark, error) {
		i := indexOf(items, path)
		if i < 0 {
			return nil, errors.New("not a bookmark: " + path)
		}
		if label == "" {
			label = labelOf(path)
		}
		items[i].Label = label
		return items, nil
	})
}

func (s *Store) Reorder(paths []string) ([]Bookmark, error) {
	return s.update(func(items []Bookmark) ([]Bookmark, error) {
		if len(paths) != len(items) {
			return nil, errors.New("reorder must list every bookmark")
		}
		next := make([]Bookmark, 0, len(items))
		for _, path := range paths {
			i := indexOf(items, path)
			if i < 0 || slices.ContainsFunc(next, func(b Bookmark) bool { return b.Path == path }) {
				return nil, errors.New("reorder must list every bookmark once")
			}
			next = append(next, items[i])
		}
		return next, nil
	})
}

func (s *Store) update(apply func([]Bookmark) ([]Bookmark, error)) ([]Bookmark, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := apply(slices.Clone(s.items))
	if err != nil {
		return nil, err
	}
	if err := s.write(next); err != nil {
		return nil, err
	}
	s.items = next
	s.seeded = true
	s.publish(next)
	return slices.Clone(next), nil
}

func (s *Store) write(items []Bookmark) error {
	data, err := json.MarshalIndent(file{Version: fileVersion, Bookmarks: items}, "", " ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := atomicfile.WriteFile(s.file, data, 0o644); err != nil {
		return err
	}
	s.lastWritten = data
	s.gtkImport = false
	return nil
}

func (s *Store) load() (items []Bookmark, data []byte, seeded bool, ok bool) {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return []Bookmark{}, nil, false, true
	}
	items, seeded, err = parse(data)
	if err != nil {
		log.Warnf("bookmarks: ignoring unreadable %s: %v", s.file, err)
		return []Bookmark{}, data, false, false
	}
	return items, data, seeded, true
}

func parse(data []byte) ([]Bookmark, bool, error) {
	var items []Bookmark
	seeded := false
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
		var f file
		if err := json.Unmarshal(data, &f); err != nil {
			return nil, false, err
		}
		items = f.Bookmarks
		seeded = f.Version >= fileVersion
	} else if err := json.Unmarshal(data, &items); err != nil {
		return nil, false, err
	}
	kept := items[:0]
	for _, b := range items {
		if !validPath(b.Path) {
			continue
		}
		if b.Label == "" {
			b.Label = labelOf(b.Path)
		}
		kept = append(kept, b)
	}
	return kept, seeded, nil
}

func (s *Store) watch() {
	defer s.stopped.Done()
	reload := time.NewTimer(coalesceWindow)
	if !reload.Stop() {
		<-reload.C
	}
	armed := false

	for {
		select {
		case <-s.done:
			return
		case ev, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(ev.Name) != FileName || armed {
				continue
			}
			reload.Reset(coalesceWindow)
			armed = true
		case _, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
		case <-reload.C:
			armed = false
			s.reload()
		}
	}
}

func (s *Store) reload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, data, seeded, ok := s.load()
	s.gtkImport = s.gtkImportAvailable()
	if !ok || bytes.Equal(data, s.lastWritten) {
		return
	}
	if !seeded && s.seedPaths != nil {
		items = seed(s.seedPaths, items)
		seeded = true
	}
	s.items = items
	s.seeded = seeded
	s.lastWritten = data
	s.publish(items)
}

func (s *Store) publish(items []Bookmark) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(Topic, changed{Kind: "changed", Bookmarks: slices.Clone(items)})
}

func indexOf(items []Bookmark, path string) int {
	return slices.IndexFunc(items, func(b Bookmark) bool { return b.Path == path })
}

// Virtual places such as recent:/// sit in the list beside real folders.
func validPath(path string) bool {
	return filepath.IsAbs(path) || isVirtual(path)
}

func isVirtual(path string) bool {
	scheme, _, ok := strings.Cut(path, "://")
	return ok && scheme != ""
}

func clean(path string) string {
	if isVirtual(path) {
		return path
	}
	return filepath.Clean(path)
}

func labelOf(path string) string {
	if path == "/" {
		return "/"
	}
	if scheme, _, ok := strings.Cut(path, "://"); ok {
		return scheme
	}
	return filepath.Base(path)
}
