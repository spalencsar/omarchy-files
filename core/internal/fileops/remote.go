package fileops

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/duxlabs-dev/dux-go/files"
)

const codeDisconnected files.Code = "DISCONNECTED"

type PathResolver interface {
	ResolvePath(path string) (resolved string, remote, available bool)
}

type resolvingFS struct {
	FS
	resolver PathResolver
}

func (f resolvingFS) path(path string) (string, error) {
	resolved, remote, available := f.resolver.ResolvePath(path)
	if remote && !available {
		return path, &files.Error{Code: codeDisconnected, Path: path, Err: errors.New("the network location is disconnected")}
	}
	return resolved, nil
}

func (f resolvingFS) Lstat(path string) (os.FileInfo, error) {
	path, err := f.path(path)
	if err != nil {
		return nil, err
	}
	return f.FS.Lstat(path)
}

func (f resolvingFS) ReadDir(path string) ([]os.DirEntry, error) {
	path, err := f.path(path)
	if err != nil {
		return nil, err
	}
	return f.FS.ReadDir(path)
}

func (f resolvingFS) Readlink(path string) (string, error) {
	path, err := f.path(path)
	if err != nil {
		return "", err
	}
	return f.FS.Readlink(path)
}

func (f resolvingFS) Open(path string) (io.ReadCloser, error) {
	path, err := f.path(path)
	if err != nil {
		return nil, err
	}
	return f.FS.Open(path)
}

func (f resolvingFS) Create(path string, perm os.FileMode) (Writer, error) {
	path, err := f.path(path)
	if err != nil {
		return nil, err
	}
	return f.FS.Create(path, perm)
}

func (f resolvingFS) Mkdir(path string, perm os.FileMode) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.Mkdir(path, perm)
}

func (f resolvingFS) MkdirAll(path string, perm os.FileMode) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.MkdirAll(path, perm)
}

func (f resolvingFS) Symlink(target, path string) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.Symlink(target, path)
}

func (f resolvingFS) Rename(from, to string, noReplace bool) error {
	from, err := f.path(from)
	if err != nil {
		return err
	}
	to, err = f.path(to)
	if err != nil {
		return err
	}
	return f.FS.Rename(from, to, noReplace)
}

func (f resolvingFS) Remove(path string) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.Remove(path)
}

func (f resolvingFS) RemoveAll(path string) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.RemoveAll(path)
}

func (f resolvingFS) Chmod(path string, mode os.FileMode, recursive bool) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.Chmod(path, mode, recursive)
}

func (f resolvingFS) Lchown(path string, uid, gid int, recursive bool) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.Lchown(path, uid, gid, recursive)
}

func (f resolvingFS) Chtimes(path string, mtime time.Time) error {
	path, err := f.path(path)
	if err != nil {
		return err
	}
	return f.FS.Chtimes(path, mtime)
}
