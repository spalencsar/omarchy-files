package gitstatus

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/duxlabs-dev/dux-go/files"
)

type Entry struct {
	Path      string `json:"path"`
	Code      string `json:"code"`
	Label     string `json:"label"`
	Staged    bool   `json:"staged"`
	Unstaged  bool   `json:"unstaged"`
	Aggregate bool   `json:"aggregate,omitempty"`
}

type Result struct {
	Available  bool    `json:"available"`
	Repository bool    `json:"repository"`
	Root       string  `json:"root,omitempty"`
	Branch     string  `json:"branch,omitempty"`
	Detached   bool    `json:"detached,omitempty"`
	Changed    int     `json:"changed"`
	Staged     int     `json:"staged"`
	Unstaged   int     `json:"unstaged"`
	Entries    []Entry `json:"entries"`
}

type Service struct {
	gitPath string
}

func NewService() *Service {
	path, _ := exec.LookPath("git")
	return &Service{gitPath: path}
}

func (s *Service) Status(ctx context.Context, path string) (Result, error) {
	if s == nil || s.gitPath == "" {
		return Result{Available: false}, nil
	}
	dir, err := directoryFor(path)
	if err != nil {
		return Result{Available: true}, err
	}
	root, ok, err := s.root(ctx, dir)
	if err != nil || !ok {
		return Result{Available: true, Repository: false}, err
	}
	out, err := s.run(ctx, root, nil, "status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all")
	if err != nil {
		return Result{Available: true, Repository: true, Root: root}, err
	}
	result := parseStatus(root, out)
	result.Available = true
	result.Repository = true
	result.Root = root
	return result, nil
}

func (s *Service) Stage(ctx context.Context, path string, paths []string) (Result, error) {
	return s.mutate(ctx, path, paths, true)
}

func (s *Service) Unstage(ctx context.Context, path string, paths []string) (Result, error) {
	return s.mutate(ctx, path, paths, false)
}

func (s *Service) StageAll(ctx context.Context, path string) (Result, error) {
	return s.mutateAll(ctx, path, true)
}

func (s *Service) UnstageAll(ctx context.Context, path string) (Result, error) {
	return s.mutateAll(ctx, path, false)
}

func (s *Service) mutate(ctx context.Context, path string, paths []string, stage bool) (Result, error) {
	if s == nil || s.gitPath == "" {
		return Result{Available: false}, errors.New("git is not installed")
	}
	dir, err := directoryFor(path)
	if err != nil {
		return Result{}, err
	}
	root, ok, err := s.root(ctx, dir)
	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{}, invalid("not inside a Git repository")
	}
	relative, err := relativePaths(root, paths)
	if err != nil {
		return Result{}, err
	}
	if len(relative) == 0 {
		return s.Status(ctx, root)
	}
	stdin := []byte(strings.Join(relative, "\x00") + "\x00")
	if stage {
		_, err = s.run(ctx, root, stdin, "add", "--pathspec-from-file=-", "--pathspec-file-nul")
	} else if s.hasHead(ctx, root) {
		_, err = s.run(ctx, root, stdin, "reset", "--mixed", "--pathspec-from-file=-", "--pathspec-file-nul", "HEAD")
	} else {
		_, err = s.run(ctx, root, stdin, "rm", "--cached", "-r", "--ignore-unmatch", "--pathspec-from-file=-", "--pathspec-file-nul")
	}
	if err != nil {
		return Result{}, err
	}
	return s.Status(ctx, root)
}

func (s *Service) mutateAll(ctx context.Context, path string, stage bool) (Result, error) {
	if s == nil || s.gitPath == "" {
		return Result{Available: false}, errors.New("git is not installed")
	}
	dir, err := directoryFor(path)
	if err != nil {
		return Result{}, err
	}
	root, ok, err := s.root(ctx, dir)
	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{}, invalid("not inside a Git repository")
	}
	if stage {
		_, err = s.run(ctx, root, nil, "add", "-A")
	} else if s.hasHead(ctx, root) {
		_, err = s.run(ctx, root, nil, "reset", "--mixed", "HEAD")
	} else {
		_, err = s.run(ctx, root, nil, "rm", "--cached", "-r", "--ignore-unmatch", ".")
	}
	if err != nil {
		return Result{}, err
	}
	return s.Status(ctx, root)
}

func (s *Service) root(ctx context.Context, dir string) (string, bool, error) {
	out, err := s.run(ctx, dir, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		if ctx.Err() != nil {
			return "", false, ctx.Err()
		}
		return "", false, nil
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", false, nil
	}
	return filepath.Clean(root), true, nil
}

func (s *Service) hasHead(ctx context.Context, root string) bool {
	_, err := s.run(ctx, root, nil, "rev-parse", "--verify", "HEAD")
	return err == nil
}

func (s *Service) run(ctx context.Context, dir string, stdin []byte, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, s.gitPath, args...)
	cmd.Dir = dir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}
	message := strings.TrimSpace(stderr.String())
	if message == "" {
		message = err.Error()
	}
	return nil, errors.New(message)
}

func directoryFor(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", invalid("path must be absolute")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return filepath.Dir(path), nil
	}
	return filepath.Clean(path), nil
}

func relativePaths(root string, paths []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return nil, invalid("Git paths must be absolute")
		}
		rel, err := filepath.Rel(root, filepath.Clean(path))
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, invalid("Git path is outside the repository")
		}
		rel = filepath.ToSlash(rel)
		if !seen[rel] {
			seen[rel] = true
			out = append(out, rel)
		}
	}
	sort.Strings(out)
	return out, nil
}

func invalid(message string) error {
	return &files.Error{Code: files.CodeInvalid, Err: errors.New(message)}
}
