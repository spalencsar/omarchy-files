package ipc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spalencsar/omarchy-files/core/internal/location"
)

func TestUIOpenRejectsRelativePath(t *testing.T) {
	deps := Deps{Bus: NewEventBus(), Pending: &PendingOpen{}}
	out := routeAndRead(t, Request{ID: 1, Method: "ui.open", Params: map[string]any{"path": "docs"}}, deps)

	errMsg, ok := out["error"].(string)
	require.True(t, ok)
	assert.Contains(t, errMsg, "not an absolute path")
	assert.Empty(t, deps.Pending.Take())
}

func TestUIOpenQueuesEveryRequestUntilSubscribed(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(file, nil, 0o644))
	deps := Deps{Bus: NewEventBus(), Pending: &PendingOpen{}}

	out := routeAndRead(t, Request{ID: 2, Method: "ui.open", Params: map[string]any{"path": dir}}, deps)
	assert.Equal(t, map[string]any{"ok": true}, out["result"])
	out = routeAndRead(t, Request{ID: 3, Method: "ui.open", Params: map[string]any{"select": []any{"file://" + file}}}, deps)
	assert.Equal(t, map[string]any{"ok": true}, out["result"])

	assert.Equal(t, []map[string]any{
		{"action": "open", "locations": []location.Location{{Kind: location.KindFolder, Path: dir, Select: []string{}}}},
		{"action": "open", "locations": []location.Location{{Kind: location.KindFolder, Path: dir, Select: []string{file}}}},
	}, deps.Pending.Take())
	assert.Empty(t, deps.Pending.Take())
}

func TestUIOpenNewWindowPassesThroughAndAllowsNoLocation(t *testing.T) {
	dir := t.TempDir()
	deps := Deps{Bus: NewEventBus(), Pending: &PendingOpen{}}

	out := routeAndRead(t, Request{ID: 6, Method: "ui.open", Params: map[string]any{"path": dir, "newWindow": true}}, deps)
	assert.Equal(t, map[string]any{"ok": true}, out["result"])
	out = routeAndRead(t, Request{ID: 7, Method: "ui.open", Params: map[string]any{"newWindow": true}}, deps)
	assert.Equal(t, map[string]any{"ok": true}, out["result"])
	out = routeAndRead(t, Request{ID: 8, Method: "ui.open", Params: map[string]any{}}, deps)
	assert.Contains(t, out["error"], "nothing to open")

	assert.Equal(t, []map[string]any{
		{"action": "open", "locations": []location.Location{{Kind: location.KindFolder, Path: dir, Select: []string{}}}, "newWindow": true},
		{"action": "open", "locations": []location.Location{}, "newWindow": true},
	}, deps.Pending.Take())
}

func TestUIOpenMarksArchives(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bundle.zip")
	require.NoError(t, os.WriteFile(archive, []byte("PK\x03\x04"), 0o644))
	deps := Deps{Bus: NewEventBus(), Pending: &PendingOpen{}}

	out := routeAndRead(t, Request{ID: 9, Method: "ui.open", Params: map[string]any{"path": archive}}, deps)
	assert.Equal(t, map[string]any{"ok": true}, out["result"])
	assert.Equal(t, []map[string]any{
		{"action": "open", "locations": []location.Location{{Kind: location.KindArchive, Path: archive, Select: []string{}}}},
	}, deps.Pending.Take())
}

func TestUIPropertiesResolvesURIs(t *testing.T) {
	dir := t.TempDir()
	deps := Deps{Bus: NewEventBus(), Pending: &PendingOpen{}}

	out := routeAndRead(t, Request{ID: 4, Method: "ui.properties", Params: map[string]any{"paths": []any{"file://" + dir}}}, deps)
	assert.Equal(t, map[string]any{"ok": true}, out["result"])
	assert.Equal(t, []map[string]any{{"action": "properties", "paths": []string{dir}}}, deps.Pending.Take())

	out = routeAndRead(t, Request{ID: 5, Method: "ui.properties", Params: map[string]any{"paths": []any{filepath.Join(dir, "gone")}}}, deps)
	assert.Contains(t, out["error"], "does not exist")
}

func TestUIRequestsKeepTheirOrderUntilTheQueueIsFlushed(t *testing.T) {
	bus := NewEventBus()
	pending := &PendingOpen{}
	pending.Deliver(bus, map[string]any{"n": 1})

	_, server := net.Pipe()
	t.Cleanup(func() { _ = server.Close() })
	sub := bus.NewSubscriber(context.Background(), NewConnWriter(server))
	t.Cleanup(sub.Close)
	sub.Subscribe("ui")
	pending.Deliver(bus, map[string]any{"n": 2})

	assert.Equal(t, []map[string]any{{"n": 1}, {"n": 2}}, pending.Take(), "a request arriving between subscribe and flush queues behind the older one")
}
