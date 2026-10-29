package gitstatus

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusAndStageRoundTrip(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	runGit(t, root, "init", "-b", "master")
	runGit(t, root, "config", "user.email", "test@example.invalid")
	runGit(t, root, "config", "user.name", "Test")
	write(t, filepath.Join(root, "tracked.txt"), "one\n")
	runGit(t, root, "add", "tracked.txt")
	runGit(t, root, "commit", "-m", "initial")

	write(t, filepath.Join(root, "tracked.txt"), "two\n")
	write(t, filepath.Join(root, "folder", "new file.txt"), "new\n")
	service := NewService()
	result, err := service.Status(context.Background(), filepath.Join(root, "folder"))
	require.NoError(t, err)
	assert.True(t, result.Repository)
	assert.Equal(t, root, result.Root)
	assert.Equal(t, "master", result.Branch)
	assert.Equal(t, 2, result.Changed)
	assertEntry(t, result, filepath.Join(root, "tracked.txt"), "modified", false, true, false)
	assertEntry(t, result, filepath.Join(root, "folder", "new file.txt"), "untracked", false, true, false)
	assertEntry(t, result, filepath.Join(root, "folder"), "untracked", false, true, true)

	result, err = service.Stage(context.Background(), root, []string{filepath.Join(root, "folder", "new file.txt")})
	require.NoError(t, err)
	assertEntry(t, result, filepath.Join(root, "folder", "new file.txt"), "added", true, false, false)

	result, err = service.Unstage(context.Background(), root, []string{filepath.Join(root, "folder", "new file.txt")})
	require.NoError(t, err)
	assertEntry(t, result, filepath.Join(root, "folder", "new file.txt"), "untracked", false, true, false)
}

func TestRejectsPathOutsideRepository(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	runGit(t, root, "init", "-b", "master")
	service := NewService()
	_, err := service.Stage(context.Background(), root, []string{filepath.Join(filepath.Dir(root), "private.txt")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the repository")
}

func TestStagePathsWithArgumentLikeAndMultilineNames(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	runGit(t, root, "init", "-b", "master")
	path := filepath.Join(root, "-private\nname.txt")
	write(t, path, "private\n")

	service := NewService()
	result, err := service.Stage(context.Background(), root, []string{path})
	require.NoError(t, err)
	assertEntry(t, result, path, "added", true, false, false)

	result, err = service.Unstage(context.Background(), root, []string{path})
	require.NoError(t, err)
	assertEntry(t, result, path, "untracked", false, true, false)
}

func TestStageAllAndUnstageAllBeforeFirstCommit(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	runGit(t, root, "init", "-b", "master")
	path := filepath.Join(root, "first.txt")
	write(t, path, "first\n")

	service := NewService()
	result, err := service.StageAll(context.Background(), root)
	require.NoError(t, err)
	assertEntry(t, result, path, "added", true, false, false)

	result, err = service.UnstageAll(context.Background(), root)
	require.NoError(t, err)
	assertEntry(t, result, path, "untracked", false, true, false)
}

func TestNotARepository(t *testing.T) {
	requireGit(t)
	result, err := NewService().Status(context.Background(), t.TempDir())
	require.NoError(t, err)
	assert.True(t, result.Available)
	assert.False(t, result.Repository)
}

func assertEntry(t *testing.T, result Result, path, code string, staged, unstaged, aggregate bool) {
	t.Helper()
	for _, entry := range result.Entries {
		if entry.Path == path {
			assert.Equal(t, code, entry.Code)
			assert.Equal(t, staged, entry.Staged)
			assert.Equal(t, unstaged, entry.Unstaged)
			assert.Equal(t, aggregate, entry.Aggregate)
			return
		}
	}
	t.Fatalf("missing status entry %s", path)
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func write(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
}
