package fileops

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type remountResolver struct {
	from      string
	to        string
	available bool
	live      *atomic.Bool
}

func (r remountResolver) ResolvePath(path string) (string, bool, bool) {
	if path != r.from && !strings.HasPrefix(path, r.from+string(filepath.Separator)) {
		return path, false, true
	}
	available := r.available
	if r.live != nil {
		available = r.live.Load()
	}
	if !available {
		return path, true, false
	}
	return r.to + strings.TrimPrefix(path, r.from), true, true
}

func TestCopyRebasesDestinationAfterRemoteRemount(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	src := filepath.Join(srcDir, "SingletonCookie")
	require.NoError(t, os.Symlink("missing-cookie-target", src))
	stale := filepath.Join(root, "gvfs", "nfs:prefix=Volume")
	current := filepath.Join(root, "gvfs", "nfs:prefix=volume")
	require.NoError(t, os.MkdirAll(current, 0o755))

	manager, bus := newHarness(t)
	manager.SetPathResolver(remountResolver{from: stale, to: current, available: true})
	view := run(t, manager, bus, spec{kind: KindCopy, sources: []string{src}, dest: stale})

	require.Equal(t, StateDone, view.State)
	target, err := os.Readlink(filepath.Join(current, "SingletonCookie"))
	require.NoError(t, err)
	require.Equal(t, "missing-cookie-target", target)
}

func TestDisconnectedRemoteGetsSpecificFailureCode(t *testing.T) {
	root := t.TempDir()
	src := write(t, filepath.Join(root, "src", "a.txt"), "a")
	stale := filepath.Join(root, "gvfs", "sftp:host=gone")

	manager, bus := newHarness(t)
	manager.SetPathResolver(remountResolver{from: stale, available: false})
	bus.onFailure = func(failure Failure) Answer {
		require.Equal(t, "DISCONNECTED", failure.Code)
		return Answer{ID: failure.ID, Action: ActionCancel}
	}
	view := run(t, manager, bus, spec{kind: KindCopy, sources: []string{src}, dest: stale})

	require.Equal(t, StateDone, view.State)
	require.True(t, view.Cancelled)
}

func TestRetryContinuesAfterRemoteReconnect(t *testing.T) {
	root := t.TempDir()
	src := write(t, filepath.Join(root, "src", "a.txt"), "a")
	stale := filepath.Join(root, "gvfs", "sftp:host=gone")
	current := filepath.Join(root, "gvfs", "sftp:host=back")
	require.NoError(t, os.MkdirAll(current, 0o755))
	var live atomic.Bool

	manager, bus := newHarness(t)
	manager.SetPathResolver(remountResolver{from: stale, to: current, live: &live})
	bus.onFailure = func(failure Failure) Answer {
		require.Equal(t, "DISCONNECTED", failure.Code)
		live.Store(true)
		return Answer{ID: failure.ID, Action: ActionRetry}
	}
	view := run(t, manager, bus, spec{kind: KindCopy, sources: []string{src}, dest: stale})

	require.Equal(t, StateDone, view.State)
	require.Equal(t, "a", read(t, filepath.Join(current, "a.txt")))
}
