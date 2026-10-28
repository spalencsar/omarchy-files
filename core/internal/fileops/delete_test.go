package fileops

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type gatedDeleteFS struct {
	FS

	mu      sync.Mutex
	removes int
	paused  chan struct{}
	resume  chan struct{}
}

func (f *gatedDeleteFS) Remove(path string) error {
	f.mu.Lock()
	f.removes++
	call := f.removes
	f.mu.Unlock()
	if call == 2 {
		close(f.paused)
		<-f.resume
	}
	return f.FS.Remove(path)
}

func TestPermanentDeleteReportsProgressWithinATree(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "remote-tree")
	write(t, filepath.Join(target, "a.txt"), "a")
	write(t, filepath.Join(target, "b.txt"), "bb")

	previous := local
	gate := &gatedDeleteFS{FS: previous, paused: make(chan struct{}), resume: make(chan struct{})}
	local = gate
	t.Cleanup(func() { local = previous })

	manager, bus := newHarness(t)
	job := manager.enqueue(spec{kind: KindDelete, sources: []string{target}})
	<-gate.paused

	during := job.view()
	require.Equal(t, StateRunning, during.State)
	require.Equal(t, 1, during.Progress.FilesDone)
	require.Equal(t, int64(1), during.Progress.BytesDone)
	require.Equal(t, filepath.Join(target, "a.txt"), during.Progress.CurrentPath)

	close(gate.resume)
	finished := bus.wait(t, job.id)
	require.Equal(t, StateDone, finished.State)
	require.Equal(t, 3, finished.Progress.FilesDone)
	require.Equal(t, int64(3), finished.Progress.BytesDone)
	require.NoDirExists(t, target)
	_, err := os.Lstat(target)
	require.True(t, os.IsNotExist(err))
}
