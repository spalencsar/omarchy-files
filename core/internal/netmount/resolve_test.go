package netmount

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePathRebasesCanonicalisedGVfsMount(t *testing.T) {
	runtime := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", runtime)
	current := filepath.Join(runtime, "gvfs", "nfs:host=10.10.10.2,prefix=%2Fvolume1%2Fbackup")
	require.NoError(t, os.MkdirAll(current, 0o700))

	svc := NewService(nil, t.TempDir())
	svc.mounts = []Mount{{
		ID:   "nfs:host=10.10.10.2,prefix=%2Fvolume1%2Fbackup",
		Path: current,
	}}
	stale := filepath.Join(runtime, "gvfs", "nfs:host=10.10.10.2,prefix=%2FVolume1%2Fbackup", "Brave", "SingletonCookie")

	resolved, remote, available := svc.ResolvePath(stale)
	assert.True(t, remote)
	assert.True(t, available)
	assert.Equal(t, filepath.Join(current, "Brave", "SingletonCookie"), resolved)
}

func TestResolvePathReportsDisconnectedGVfsMount(t *testing.T) {
	runtime := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", runtime)
	stale := filepath.Join(runtime, "gvfs", "sftp:host=gone", "folder", "file")

	resolved, remote, available := NewService(nil, t.TempDir()).ResolvePath(stale)
	assert.Equal(t, stale, resolved)
	assert.True(t, remote)
	assert.False(t, available)
}

func TestResolvePathLeavesLocalPathsAlone(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	resolved, remote, available := NewService(nil, t.TempDir()).ResolvePath("/home/me/file")
	assert.Equal(t, "/home/me/file", resolved)
	assert.False(t, remote)
	assert.True(t, available)
}
