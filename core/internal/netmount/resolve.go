package netmount

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath maps a GVfs FUSE path to the currently mounted path for the
// same remote. GVfs may recreate a mount with a differently-canonicalised
// mountpoint (for example changing the case of an NFS prefix).
//
// remote reports whether path belongs to the GVfs mount directory. available
// is false when that remote is no longer mounted.
func (s *Service) ResolvePath(path string) (resolved string, remote, available bool) {
	root, rest, ok := gvfsPath(path)
	if !ok {
		return path, false, true
	}

	s.mu.RLock()
	mounts := append([]Mount(nil), s.mounts...)
	s.mu.RUnlock()

	oldID := filepath.Base(root)
	for _, mount := range mounts {
		if mount.Path == "" || !mountPathExists(mount.Path) {
			continue
		}
		currentRoot := filepath.Clean(mount.Path)
		if root == currentRoot || strings.EqualFold(oldID, mount.ID) || strings.EqualFold(oldID, filepath.Base(currentRoot)) {
			if rest == "" {
				return currentRoot, true, true
			}
			return filepath.Join(currentRoot, rest), true, true
		}
	}
	return path, true, false
}

func gvfsPath(path string) (root, rest string, ok bool) {
	runtime := os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		return "", "", false
	}
	base := filepath.Join(runtime, "gvfs")
	relative, err := filepath.Rel(base, filepath.Clean(path))
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", false
	}
	parts := strings.SplitN(relative, string(filepath.Separator), 2)
	root = filepath.Join(base, parts[0])
	if len(parts) == 2 {
		rest = parts[1]
	}
	return root, rest, true
}

var mountPathExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
