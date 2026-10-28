package privatefiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardenCreatesPrivateTreeAndRepairsExistingModes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "omarchy-files")
	nested := filepath.Join(root, "thumbnails")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	file := filepath.Join(nested, "private.json")
	require.NoError(t, os.WriteFile(file, []byte("secret"), 0o644))

	require.NoError(t, Harden(root))

	for _, path := range []string{root, nested} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(), path)
	}
	info, err := os.Stat(file)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestHardenDoesNotFollowSymlinks(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "omarchy-files")
	outside := filepath.Join(base, "outside")
	require.NoError(t, os.MkdirAll(root, 0o755))
	require.NoError(t, os.WriteFile(outside, []byte("public"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "link")))

	require.NoError(t, Harden(root))

	info, err := os.Stat(outside)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}
