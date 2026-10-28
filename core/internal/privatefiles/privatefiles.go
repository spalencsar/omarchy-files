// Package privatefiles protects Omarchy Files' app-owned storage from other
// local users. The stored session, folder history, bookmarks, server URIs and
// thumbnails can all disclose private filesystem locations.
package privatefiles

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Harden recursively restricts an app-owned directory to its owner. Symlinks
// are deliberately not followed or changed.
func Harden(root string) error {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return os.Chmod(path, 0o700)
		}
		if entry.Type().IsRegular() {
			return os.Chmod(path, 0o600)
		}
		return nil
	})
}
