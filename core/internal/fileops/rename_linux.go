package fileops

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

var renameAt2 = unix.Renameat2

func renameNoReplace(from, to string) error {
	err := renameAt2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
	if err == nil {
		return nil
	}
	// Some FUSE, network and otherwise local-looking filesystems implement
	// rename(2) but reject renameat2's RENAME_NOREPLACE flag with EINVAL,
	// ENOSYS or EOPNOTSUPP. The destination has already been conflict-checked;
	// check it again before using plain rename as a compatibility fallback.
	if !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.EOPNOTSUPP) {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	if _, statErr := os.Lstat(to); statErr == nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: os.ErrExist}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: statErr}
	}
	if renameErr := os.Rename(from, to); renameErr != nil {
		return renameErr
	}
	return nil
}
