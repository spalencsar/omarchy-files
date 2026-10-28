package ipc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/duxlabs-dev/dux-go/files"
	"github.com/spalencsar/omarchy-files/core/internal/recent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useRecentStore(t *testing.T) (*recent.Store, string) {
	t.Helper()
	root := t.TempDir()
	return recent.NewStore(filepath.Join(root, "data")), root
}

func TestRecentEntriesListsUsableFilesNewestFirst(t *testing.T) {
	store, root := useRecentStore(t)
	for _, name := range []string{"old.txt", "new.txt", "hidden.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte("x"), 0o600))
	}
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	require.NoError(t, store.Add(filepath.Join(root, "old.txt"), "text/plain", "App", "'gio open %u'", base))
	require.NoError(t, store.Add(filepath.Join(root, "gone.txt"), "text/plain", "App", "'gio open %u'", base.Add(time.Minute)))
	require.NoError(t, store.Add(filepath.Join(root, "new.txt"), "text/plain", "App", "'gio open %u'", base.Add(2*time.Minute)))
	require.NoError(t, store.Add(filepath.Join(root, "hidden.txt"), "text/plain", "App", "'gio open %u'", base.Add(3*time.Minute)))
	items, _, err := store.Load()
	require.NoError(t, err)
	for i := range items {
		if items[i].URI == recent.FileURI(filepath.Join(root, "hidden.txt")) {
			items[i].Private = true
		}
	}
	items = append(items, recent.Item{URI: "https://example.com/doc", Modified: base.Add(time.Hour)})
	require.NoError(t, os.WriteFile(store.Path, recent.Marshal(items), 0o600))

	out := routeAndRead(t, Request{ID: 3, Method: "recent.entries"}, Deps{Recent: store, Files: files.NewService(nil, t.TempDir(), nil)})
	result, ok := out["result"].(map[string]any)
	require.True(t, ok, out)
	entries, ok := result["entries"].([]any)
	require.True(t, ok)
	require.Len(t, entries, 2)
	newest := entries[0].(map[string]any)
	assert.Equal(t, "new.txt", newest["name"])
	assert.Equal(t, float64(base.Add(2*time.Minute).UnixMilli()), newest["usedMs"])
	assert.Equal(t, recent.FileURI(filepath.Join(root, "new.txt")), newest["uri"])
	assert.Equal(t, "old.txt", entries[1].(map[string]any)["name"])
	assert.Equal(t, store.Stamp(), result["stamp"])

	again := routeAndRead(t, Request{ID: 4, Method: "recent.entries", Params: map[string]any{"stamp": result["stamp"]}}, Deps{Recent: store, Files: files.NewService(nil, t.TempDir(), nil)})
	assert.Equal(t, true, again["result"].(map[string]any)["unchanged"])
}

func TestRecentRemoveAndClearChangeTheStamp(t *testing.T) {
	store, root := useRecentStore(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))
	require.NoError(t, store.Add(filepath.Join(root, "a.txt"), "text/plain", "App", "'gio open %u'", time.Now()))
	require.NoError(t, store.Add(filepath.Join(root, "b.txt"), "text/plain", "App", "'gio open %u'", time.Now()))
	before := store.Stamp()

	out := routeAndRead(t, Request{ID: 5, Method: "recent.remove", Params: map[string]any{"paths": []any{filepath.Join(root, "a.txt")}}}, Deps{Recent: store})
	result := out["result"].(map[string]any)
	assert.Equal(t, true, result["ok"])
	assert.NotEqual(t, before, result["stamp"])
	items, _, err := store.Load()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, recent.FileURI(filepath.Join(root, "b.txt")), items[0].URI)

	bad := routeAndRead(t, Request{ID: 6, Method: "recent.remove"}, Deps{Recent: store})
	assert.NotNil(t, bad["error"])

	out = routeAndRead(t, Request{ID: 7, Method: "recent.clear"}, Deps{Recent: store})
	assert.Equal(t, true, out["result"].(map[string]any)["ok"])
	items, _, err = store.Load()
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestRecentEntriesWithoutAFileService(t *testing.T) {
	store, _ := useRecentStore(t)
	out := routeAndRead(t, Request{ID: 8, Method: "recent.entries"}, Deps{Recent: store})
	assert.NotNil(t, out["error"])
	out = routeAndRead(t, Request{ID: 9, Method: "recent.entries"}, Deps{Files: files.NewService(nil, t.TempDir(), nil)})
	assert.NotNil(t, out["error"])
}
