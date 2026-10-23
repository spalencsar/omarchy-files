package elevate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/duxlabs-dev/dux-go/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spalencsar/omarchy-files/core/internal/fileops"
	"github.com/spalencsar/omarchy-files/core/internal/helper"
)

const (
	fakeEnv     = "OMARCHY_FILES_ELEVATE_FAKE_HELPER"
	fakeExitEnv = "OMARCHY_FILES_ELEVATE_FAKE_EXIT"
)

// TestMain lets the test binary stand in for pkexec: invoked as
// "<binary> <exe> helper <op> ..." it runs the helper operation unprivileged.
func TestMain(m *testing.M) {
	if code := os.Getenv(fakeExitEnv); code != "" {
		n, _ := strconv.Atoi(code)
		os.Exit(n)
	}
	if os.Getenv(fakeEnv) == "" {
		os.Exit(m.Run())
	}
	args := os.Args[1:]
	if len(args) < 2 || args[1] != "helper" {
		fmt.Fprintln(os.Stderr, "bad fake invocation")
		os.Exit(127)
	}
	if err := helper.Run(args[2:], os.Stdin, os.Stdout); err != nil {
		line, _ := json.Marshal(helper.Failure{Code: string(helper.CodeOf(err)), Error: err.Error()})
		fmt.Fprintln(os.Stderr, string(line))
		os.Exit(helper.ExitFailed)
	}
	os.Exit(0)
}

func fakeClient(t *testing.T) *Client {
	t.Helper()
	self, err := os.Executable()
	require.NoError(t, err)
	return &Client{pkexec: self, exe: "omarchy-files", env: []string{fakeEnv + "=1"}}
}

func TestPolicyCoversOnlyThisBinary(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "bin", "omarchy-files")
	require.NoError(t, os.MkdirAll(filepath.Dir(exe), 0o755))
	require.NoError(t, os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755))
	link := filepath.Join(dir, "omarchy-files-link")
	require.NoError(t, os.Symlink(exe, link))

	actions := filepath.Join(dir, "actions")
	require.NoError(t, os.Mkdir(actions, 0o755))
	policy := `<policyconfig><action id="` + ActionID + `"><annotate key="org.freedesktop.policykit.exec.path">` + link + `</annotate></action></policyconfig>`
	require.NoError(t, os.WriteFile(filepath.Join(actions, PolicyName), []byte(policy), 0o644))

	assert.True(t, policyCovers(exe, []string{"/nonexistent", actions}), "the annotation is resolved through symlinks")
	assert.False(t, policyCovers(filepath.Join(dir, "other"), []string{actions}))
	assert.False(t, policyCovers(exe, []string{"/nonexistent"}))
	assert.Equal(t, "", annotatedPath("<policyconfig/>"))
}

func TestFailureMapsExitCodesAndHelperReports(t *testing.T) {
	self, err := os.Executable()
	require.NoError(t, err)
	exitWith := func(code int) error {
		cmd := exec.Command(self)
		cmd.Env = append(os.Environ(), fakeExitEnv+"="+strconv.Itoa(code))
		return cmd.Run()
	}

	err = failure(exitWith(126), nil)
	assert.Equal(t, files.CodeAccess, files.CodeOf(err))
	assert.Contains(t, err.Error(), "not granted")

	err = failure(exitWith(127), []byte("Error executing command as another user: No authentication agent\n"))
	assert.Equal(t, files.CodeNotSupported, files.CodeOf(err))
	assert.Contains(t, err.Error(), "No authentication agent")

	err = failure(exitWith(127), []byte("Error executing command as another user: Not authorized\n"))
	assert.Equal(t, files.CodeAccess, files.CodeOf(err))

	err = failure(exitWith(1), []byte("noise\n{\"code\":\"ENOENT\",\"error\":\"lstat /x: no such file\"}\n"))
	assert.Equal(t, files.CodeNotFound, files.CodeOf(err))
	assert.Equal(t, "ENOENT: lstat /x: no such file", err.Error())

	err = failure(exitWith(3), []byte("garbage"))
	assert.Equal(t, files.CodeIO, files.CodeOf(err))

	err = failure(errors.New("fork failed"), nil)
	assert.Equal(t, files.CodeIO, files.CodeOf(err))
}

func TestClientSendsPermissionBitsOnly(t *testing.T) {
	root := t.TempDir()
	client := fakeClient(t)

	dir := filepath.Join(root, "shared")
	require.NoError(t, client.Mkdir(dir, 0o2775))
	require.NoError(t, client.Chmod(dir, 0o755|os.ModeSetgid|os.ModeSticky, false))
	info, err := os.Lstat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	assert.Zero(t, info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky))
	assert.Equal(t, "755", octal(0o755|os.ModeSetuid))
}

func TestClientRoundTripsThroughTheHelper(t *testing.T) {
	root := t.TempDir()
	client := fakeClient(t)

	require.NoError(t, client.MkdirAll(filepath.Join(root, "a", "b"), 0o755))
	require.NoError(t, client.Mkdir(filepath.Join(root, "a", "c"), 0o700))

	out, err := client.Create(filepath.Join(root, "a", "f.txt"), 0o640)
	require.NoError(t, err)
	_, err = out.Write([]byte("hello "))
	require.NoError(t, err)
	_, err = out.Write(bytes.Repeat([]byte("w"), 100000))
	require.NoError(t, err)
	require.NoError(t, out.Close())
	require.NoError(t, out.Close(), "closing twice is harmless")

	info, err := client.Lstat(filepath.Join(root, "a", "f.txt"))
	require.NoError(t, err)
	assert.Equal(t, "f.txt", info.Name())
	assert.Equal(t, int64(100006), info.Size())
	assert.True(t, info.Mode().IsRegular())
	assert.False(t, info.ModTime().IsZero())

	in, err := client.Open(filepath.Join(root, "a", "f.txt"))
	require.NoError(t, err)
	body, err := io.ReadAll(in)
	require.NoError(t, err)
	require.NoError(t, in.Close())
	assert.Len(t, body, 100006)
	assert.Equal(t, "hello ", string(body[:6]))

	require.NoError(t, client.Symlink("f.txt", filepath.Join(root, "a", "l")))
	target, err := client.Readlink(filepath.Join(root, "a", "l"))
	require.NoError(t, err)
	assert.Equal(t, "f.txt", target)

	entries, err := client.ReadDir(filepath.Join(root, "a"))
	require.NoError(t, err)
	names := map[string]bool{}
	for _, entry := range entries {
		names[entry.Name()] = entry.IsDir()
	}
	assert.Equal(t, map[string]bool{"b": true, "c": true, "f.txt": false, "l": false}, names)

	require.NoError(t, client.Rename(filepath.Join(root, "a", "f.txt"), filepath.Join(root, "a", "g.txt"), true))
	require.NoError(t, os.WriteFile(filepath.Join(root, "a", "h.txt"), []byte("h"), 0o644))
	err = client.Rename(filepath.Join(root, "a", "g.txt"), filepath.Join(root, "a", "h.txt"), true)
	assert.Equal(t, files.CodeExists, files.CodeOf(err), "noreplace refuses an existing target")
	require.NoError(t, client.Rename(filepath.Join(root, "a", "g.txt"), filepath.Join(root, "a", "h.txt"), false))
	stamp := time.Unix(1700000000, 0)
	require.NoError(t, client.Chtimes(filepath.Join(root, "a", "h.txt"), stamp))
	stamped, err := os.Lstat(filepath.Join(root, "a", "h.txt"))
	require.NoError(t, err)
	assert.True(t, stamped.ModTime().Equal(stamp))
	require.NoError(t, client.Chmod(filepath.Join(root, "a"), 0o700, true))
	require.NoError(t, client.Lchown(filepath.Join(root, "a", "h.txt"), -1, -1, false))
	require.NoError(t, client.Remove(filepath.Join(root, "a", "l")))
	require.NoError(t, client.RemoveAll(filepath.Join(root, "a")))
	_, err = os.Lstat(filepath.Join(root, "a"))
	assert.True(t, os.IsNotExist(err))
}

func TestClientSurfacesHelperErrors(t *testing.T) {
	root := t.TempDir()
	client := fakeClient(t)

	_, err := client.Lstat(filepath.Join(root, "missing"))
	assert.Equal(t, files.CodeNotFound, files.CodeOf(err))

	_, err = client.Open(filepath.Join(root, "missing"))
	assert.Equal(t, files.CodeNotFound, files.CodeOf(err), "a refused open is reported by Open itself")

	empty := filepath.Join(root, "empty")
	require.NoError(t, os.WriteFile(empty, nil, 0o644))
	in, err := client.Open(empty)
	require.NoError(t, err)
	body, err := io.ReadAll(in)
	require.NoError(t, err)
	assert.Empty(t, body)
	require.NoError(t, in.Close())

	existing := filepath.Join(root, "exists")
	require.NoError(t, os.WriteFile(existing, []byte("old"), 0o644))
	_, err = client.Create(existing, 0o644)
	assert.Equal(t, files.CodeExists, files.CodeOf(err), "a refused create is reported by Create itself")
	body, _ = os.ReadFile(existing)
	assert.Equal(t, "old", string(body))

	aborted := filepath.Join(root, "aborted")
	out, err := client.Create(aborted, 0o644)
	require.NoError(t, err)
	_, err = out.Write([]byte("partial"))
	require.NoError(t, err)
	require.NoError(t, out.Abort())
	_, statErr := os.Lstat(aborted)
	assert.True(t, os.IsNotExist(statErr), "an aborted write leaves nothing behind")

	err = client.Mkdir("relative/dir", 0o755)
	assert.Equal(t, files.CodeInvalid, files.CodeOf(err))
}

func TestClientCanonicalisesSymlinkedParents(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	require.NoError(t, os.Mkdir(real, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(real, "f"), []byte("f"), 0o644))
	link := filepath.Join(root, "link")
	require.NoError(t, os.Symlink(real, link))
	client := fakeClient(t)

	info, err := client.Lstat(filepath.Join(link, "f"))
	require.NoError(t, err)
	assert.Equal(t, "f", info.Name())

	require.NoError(t, client.Mkdir(filepath.Join(link, "sub"), 0o755))
	_, err = os.Lstat(filepath.Join(real, "sub"))
	assert.NoError(t, err)

	assert.Equal(t, filepath.Join(real, "f"), canonical(filepath.Join(link, "f")))
	assert.Equal(t, "/nonexistent/x/y", canonical("/nonexistent/x/y"), "an unresolvable parent is passed through")
}

func TestAvailableExplainsWhatIsMissing(t *testing.T) {
	t.Setenv("XDG_DATA_DIRS", t.TempDir())
	t.Setenv("FLATPAK_ID", "com.example.Sandbox")
	client := &Client{pkexec: "/bin/true", exe: "/bin/true"}
	assert.Equal(t, ReasonSandboxed, reason(t, client.Available()))

	t.Setenv("FLATPAK_ID", "")
	if _, err := os.Stat("/.flatpak-info"); err == nil {
		t.Skip("running inside flatpak")
	}
	assert.Equal(t, ReasonNoPolicy, reason(t, client.Available()))

	missing := &Client{pkexec: "/nonexistent/pkexec", exe: "/bin/true"}
	assert.Equal(t, ReasonNoPkexec, reason(t, missing.Available()))
}

func reason(t *testing.T, err error) string {
	t.Helper()
	var unavailable *fileops.Unavailable
	require.ErrorAs(t, err, &unavailable)
	return unavailable.Reason
}
