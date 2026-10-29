package gitstatus

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStatusKinds(t *testing.T) {
	root := filepath.Clean("/repo")
	data := []byte("# branch.head feature/git\x00" +
		"1 M. N... 100644 100644 100644 aaaaaaa bbbbbbb staged.txt\x00" +
		"1 .D N... 100644 100644 000000 aaaaaaa aaaaaaa deleted.txt\x00" +
		"2 R. N... 100644 100644 100644 aaaaaaa bbbbbbb R100 renamed.txt\x00old.txt\x00" +
		"u UU N... 100644 100644 100644 100644 aaaaaaa bbbbbbb ccccccc conflicted.txt\x00" +
		"? folder/new.txt\x00")

	result := parseStatus(root, data)

	assert.Equal(t, "feature/git", result.Branch)
	assert.Equal(t, 5, result.Changed)
	assert.Equal(t, 3, result.Staged)
	assert.Equal(t, 3, result.Unstaged)
	assertEntry(t, result, filepath.Join(root, "staged.txt"), "modified", true, false, false)
	assertEntry(t, result, filepath.Join(root, "deleted.txt"), "deleted", false, true, false)
	assertEntry(t, result, filepath.Join(root, "renamed.txt"), "renamed", true, false, false)
	assertEntry(t, result, filepath.Join(root, "conflicted.txt"), "conflict", true, true, false)
	assertEntry(t, result, filepath.Join(root, "folder", "new.txt"), "untracked", false, true, false)
	assertEntry(t, result, filepath.Join(root, "folder"), "untracked", false, true, true)
}

func TestFolderAggregateKeepsBothIndexStates(t *testing.T) {
	root := filepath.Clean("/repo")
	data := []byte("1 A. N... 000000 100644 100644 aaaaaaa bbbbbbb folder/staged.txt\x00" +
		"1 .M N... 100644 100644 100644 aaaaaaa bbbbbbb folder/unstaged.txt\x00")

	result := parseStatus(root, data)

	assertEntry(t, result, filepath.Join(root, "folder"), "added", true, true, true)
}
