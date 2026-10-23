package bookmarks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBus struct {
	mu     sync.Mutex
	events []changed
	notify chan changed
}

func newBus() *testBus { return &testBus{notify: make(chan changed, 16)} }

func (b *testBus) Publish(topic string, data any) {
	ev, ok := data.(changed)
	if !ok || topic != Topic {
		return
	}
	b.mu.Lock()
	b.events = append(b.events, ev)
	b.mu.Unlock()
	b.notify <- ev
}

func (b *testBus) next(t *testing.T) changed {
	t.Helper()
	select {
	case ev := <-b.notify:
		return ev
	case <-time.After(5 * time.Second):
		t.Fatal("no bookmarks event")
		return changed{}
	}
}

func newStore(t *testing.T) (*Store, *testBus, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "omarchy-files")
	bus := newBus()
	store := New(bus, dir, filepath.Join(filepath.Dir(dir), "gtk-3.0", "bookmarks"))
	require.NoError(t, store.Start())
	t.Cleanup(store.Close)
	return store, bus, dir
}

func paths(items []Bookmark) []string {
	out := make([]string, 0, len(items))
	for _, b := range items {
		out = append(out, b.Path)
	}
	return out
}

func TestAddRemoveRenameReorderRoundTrip(t *testing.T) {
	store, bus, dir := newStore(t)

	items, err := store.Add(Bookmark{Path: "/srv/media"})
	require.NoError(t, err)
	assert.Equal(t, []Bookmark{{Path: "/srv/media", Label: "media"}}, items)
	assert.Equal(t, items, bus.next(t).Bookmarks)

	_, err = store.Add(Bookmark{Path: "/home/me/Projects/", Label: "Work", Icon: "folder-code"})
	require.NoError(t, err)
	bus.next(t)
	_, err = store.Add(Bookmark{Path: "/srv/media"})
	require.NoError(t, err)
	bus.next(t)
	assert.Equal(t, []string{"/srv/media", "/home/me/Projects"}, paths(store.List()), "a duplicate add is a no-op")

	items, err = store.Rename("/srv/media", "Media")
	require.NoError(t, err)
	assert.Equal(t, "Media", items[0].Label)
	bus.next(t)

	items, err = store.Reorder([]string{"/home/me/Projects", "/srv/media"})
	require.NoError(t, err)
	assert.Equal(t, []string{"/home/me/Projects", "/srv/media"}, paths(items))
	bus.next(t)

	_, err = store.Reorder([]string{"/srv/media"})
	assert.Error(t, err, "a partial order is refused")
	_, err = store.Reorder([]string{"/srv/media", "/srv/media"})
	assert.Error(t, err, "a repeated path is refused")
	_, err = store.Add(Bookmark{Path: "relative"})
	assert.Error(t, err)

	items, err = store.Remove("/srv/media")
	require.NoError(t, err)
	assert.Equal(t, []string{"/home/me/Projects"}, paths(items))
	bus.next(t)

	data, err := os.ReadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	var stored file
	require.NoError(t, json.Unmarshal(data, &stored))
	assert.Equal(t, file{Version: 2, Bookmarks: []Bookmark{{Path: "/home/me/Projects", Label: "Work", Icon: "folder-code"}}}, stored)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "no temp file left behind")

	reopened := New(nil, dir, "")
	require.NoError(t, reopened.Start())
	t.Cleanup(reopened.Close)
	assert.Equal(t, stored.Bookmarks, reopened.List())
	assert.True(t, reopened.Seeded(), "a version 2 file counts as seeded")
}

func TestSeedFrontsTheListOnceAndStaysInMemory(t *testing.T) {
	store, bus, dir := newStore(t)
	assert.False(t, store.Seeded())

	items, err := store.Seed([]string{"recent:///", "/home/me", "/home/me/Downloads/", "/home/me"})
	require.NoError(t, err)
	assert.Equal(t, []Bookmark{{Path: "recent:///", Label: "recent"}, {Path: "/home/me", Label: "me"}, {Path: "/home/me/Downloads", Label: "Downloads"}}, items)
	assert.Equal(t, items, bus.next(t).Bookmarks)
	assert.True(t, store.Seeded())
	assert.NoFileExists(t, filepath.Join(dir, FileName), "seeding alone writes nothing")

	items, err = store.Seed([]string{"/elsewhere"})
	require.NoError(t, err)
	assert.Equal(t, []string{"recent:///", "/home/me", "/home/me/Downloads"}, paths(items), "a second seed is a no-op")

	_, err = store.Seed([]string{"relative"})
	assert.Error(t, err)
	_, err = store.Seed([]string{"://nope"})
	assert.Error(t, err)

	items, err = store.Remove("/home/me/Downloads")
	require.NoError(t, err)
	bus.next(t)
	assert.Equal(t, []string{"recent:///", "/home/me"}, paths(items))

	reopened := New(nil, dir, "")
	require.NoError(t, reopened.Start())
	t.Cleanup(reopened.Close)
	assert.True(t, reopened.Seeded(), "the first edit writes a seeded file")
	items, err = reopened.Seed([]string{"/home/me", "/home/me/Downloads"})
	require.NoError(t, err)
	assert.Equal(t, []string{"recent:///", "/home/me"}, paths(items), "a removed place stays removed")
}

func TestSeedGoesInFrontOfALegacyFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "omarchy-files")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte(`[{"path":"/srv/media","label":"Media","icon":""},{"path":"/home/me","label":"Mine","icon":"x"}]`), 0o644))
	bus := newBus()
	store := New(bus, dir, "")
	require.NoError(t, store.Start())
	t.Cleanup(store.Close)
	assert.False(t, store.Seeded(), "a bare array predates the places")

	items, err := store.Seed([]string{"/home/me", "/home/me/Documents"})
	require.NoError(t, err)
	assert.Equal(t, []Bookmark{
		{Path: "/home/me", Label: "Mine", Icon: "x"},
		{Path: "/home/me/Documents", Label: "Documents"},
		{Path: "/srv/media", Label: "Media"},
	}, items, "a place already saved keeps its label and icon")
	bus.next(t)

	external := []byte(`[{"path":"/two","label":"Two","icon":""}]`)
	tmp := filepath.Join(dir, "edit.tmp")
	require.NoError(t, os.WriteFile(tmp, external, 0o644))
	require.NoError(t, os.Rename(tmp, filepath.Join(dir, FileName)))
	assert.Equal(t, []string{"/home/me", "/home/me/Documents", "/two"}, paths(bus.next(t).Bookmarks), "a legacy file dropped in later is seeded again")
	assert.True(t, store.Seeded())
}

func TestExternalEditIsPickedUp(t *testing.T) {
	store, bus, dir := newStore(t)
	_, err := store.Add(Bookmark{Path: "/one"})
	require.NoError(t, err)
	bus.next(t)

	external := []byte(`[{"path":"/two","label":"Two","icon":""}]`)
	tmp := filepath.Join(dir, "edit.tmp")
	require.NoError(t, os.WriteFile(tmp, external, 0o644))
	require.NoError(t, os.Rename(tmp, filepath.Join(dir, FileName)))

	ev := bus.next(t)
	assert.Equal(t, []Bookmark{{Path: "/two", Label: "Two"}}, ev.Bookmarks)
	assert.Equal(t, ev.Bookmarks, store.List())
}

func TestOwnWritesDoNotEchoThroughTheWatch(t *testing.T) {
	store, bus, _ := newStore(t)
	_, err := store.Add(Bookmark{Path: "/one"})
	require.NoError(t, err)
	bus.next(t)

	select {
	case ev := <-bus.notify:
		t.Fatalf("unexpected second event: %v", ev)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestUnreadableFileLoadsEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "omarchy-files")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte("{not json"), 0o644))
	store := New(nil, dir, "")
	require.NoError(t, store.Start())
	t.Cleanup(store.Close)
	assert.Empty(t, store.List())

	items, err := store.Add(Bookmark{Path: "/fresh"})
	require.NoError(t, err)
	assert.Equal(t, []string{"/fresh"}, paths(items))
}

func TestParseGTK(t *testing.T) {
	data := []byte("file:///home/me/Documents\nfile:///home/me/My%20Music Tunes\n\nsftp://host/share Remote\nfile:///home/me/Videos  \n")
	assert.Equal(t, []Bookmark{
		{Path: "/home/me/Documents"},
		{Path: "/home/me/My Music", Label: "Tunes"},
		{Path: "/home/me/Videos"},
	}, ParseGTK(data))
}

func TestGTKImportIsOfferedOnceAndDedupes(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "omarchy-files")
	gtk := filepath.Join(filepath.Dir(dir), "gtk-3.0", "bookmarks")
	require.NoError(t, os.MkdirAll(filepath.Dir(gtk), 0o755))
	require.NoError(t, os.WriteFile(gtk, []byte("file:///a/b Bee\nfile:///c\n"), 0o644))
	bus := newBus()
	store := New(bus, dir, gtk)
	require.NoError(t, store.Start())
	t.Cleanup(store.Close)
	assert.True(t, store.GTKImportAvailable())

	items, added, err := store.ImportGTK()
	require.NoError(t, err)
	assert.Equal(t, 2, added)
	assert.Equal(t, []Bookmark{{Path: "/a/b", Label: "Bee"}, {Path: "/c", Label: "c"}}, items)
	bus.next(t)
	assert.False(t, store.GTKImportAvailable(), "once bookmarks.json exists the offer is gone")

	_, added, err = store.ImportGTK()
	require.NoError(t, err)
	assert.Equal(t, 0, added)
}

func TestGTKImportUnavailableWithoutTheFile(t *testing.T) {
	store, _, _ := newStore(t)
	assert.False(t, store.GTKImportAvailable())
	_, _, err := store.ImportGTK()
	assert.Error(t, err)
}
