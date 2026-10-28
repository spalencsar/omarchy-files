package filechooser

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/spalencsar/omarchy-files/core/internal/dbustest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const handle = dbus.ObjectPath("/org/freedesktop/portal/desktop/request/1_7/t1")

type fakeUI struct {
	shown  chan Request
	closed chan string
}

func newFakeUI() *fakeUI {
	return &fakeUI{shown: make(chan Request, 4), closed: make(chan string, 4)}
}

func (f *fakeUI) ShowChooser(req Request) { f.shown <- req }

func (f *fakeUI) CloseChooser(id string) { f.closed <- id }

func (f *fakeUI) next(t *testing.T) Request {
	t.Helper()
	select {
	case req := <-f.shown:
		return req
	case <-time.After(5 * time.Second):
		t.Fatal("no chooser request reached the ui")
		return Request{}
	}
}

func (f *fakeUI) closedID(t *testing.T) string {
	t.Helper()
	select {
	case id := <-f.closed:
		return id
	case <-time.After(5 * time.Second):
		t.Fatal("the ui was not asked to close")
		return ""
	}
}

func globs(mime string) []string {
	switch mime {
	case "image/png":
		return []string{"*.png"}
	case "image/*":
		return []string{"*.jpg", "*.png"}
	}
	return nil
}

func start(t *testing.T, address string, ui UI) *Service {
	t.Helper()
	conn, err := dbus.Connect(address)
	require.NoError(t, err)
	svc, err := New(conn, ui, globs)
	require.NoError(t, err)
	t.Cleanup(svc.Close)
	status := svc.SetEnabled(true)
	require.True(t, status.Owned)
	return svc
}

type callResult struct {
	response uint32
	results  map[string]dbus.Variant
	err      error
}

func call(conn *dbus.Conn, method string, h dbus.ObjectPath, title string, options map[string]dbus.Variant) <-chan callResult {
	out := make(chan callResult, 1)
	go func() {
		var r callResult
		r.err = conn.Object(BusName, Path).Call(Interface+"."+method, 0, h, "", "", title, options).Store(&r.response, &r.results)
		out <- r
	}()
	return out
}

func wait(t *testing.T, ch <-chan callResult) callResult {
	t.Helper()
	select {
	case r := <-ch:
		return r
	case <-time.After(5 * time.Second):
		t.Fatal("the portal call did not return")
		return callResult{}
	}
}

func bytesOption(path string) dbus.Variant {
	return dbus.MakeVariant(append([]byte(path), 0))
}

func TestOpenFileRoundTrip(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	svc := start(t, address, ui)
	client := dbustest.Connect(t, address)
	folder := t.TempDir()

	options := map[string]dbus.Variant{
		"multiple":     dbus.MakeVariant(true),
		"accept_label": dbus.MakeVariant("_Pick"),
		"filters": dbus.MakeVariant([]wireFilter{
			{Name: "Images", Patterns: []wirePattern{{Kind: 1, Pattern: "image/*"}, {Kind: 0, Pattern: "*.GIF"}}},
			{Name: "PNG", Patterns: []wirePattern{{Kind: 1, Pattern: "image/png"}}},
		}),
		"current_filter": dbus.MakeVariant(wireFilter{Name: "PNG", Patterns: []wirePattern{{Kind: 1, Pattern: "image/png"}}}),
		"choices": dbus.MakeVariant([]wireChoice{
			{ID: "encoding", Label: "Encoding", Options: []wireOption{{ID: "utf8", Label: "UTF-8"}, {ID: "latin1", Label: "Latin 1"}}, Initial: "latin1"},
			{ID: "readonly", Label: "Open read-only", Options: []wireOption{}, Initial: "true"},
		}),
		"current_folder": bytesOption(folder),
	}
	result := call(client, "OpenFile", handle, "  Open something  ", options)

	req := ui.next(t)
	assert.Equal(t, string(handle), req.ID)
	assert.Equal(t, MethodOpen, req.Method)
	assert.Equal(t, "Open something", req.Title)
	assert.Equal(t, "Pick", req.AcceptLabel)
	assert.True(t, req.Multiple)
	assert.False(t, req.Directory)
	assert.Equal(t, folder, req.CurrentFolder)
	require.Len(t, req.Filters, 2)
	assert.Equal(t, []string{"*.jpg", "*.png", "*.gif"}, req.Filters[0].Globs)
	assert.Equal(t, []Pattern{{Kind: KindMime, Pattern: "image/*"}, {Kind: KindGlob, Pattern: "*.GIF"}}, req.Filters[0].Patterns)
	assert.Equal(t, 1, req.CurrentFilter)
	require.Len(t, req.Choices, 2)
	assert.Equal(t, "latin1", req.Choices[0].Initial)
	assert.Empty(t, req.Choices[1].Options)
	require.Len(t, svc.Pending(), 1)
	assert.Equal(t, string(handle), svc.Pending()[0].ID)

	chosen := filepath.Join(folder, "a b.png")
	require.True(t, svc.Answer(req.ID, Answer{Response: ResponseOK, Paths: []string{chosen, "relative.png"}, Choices: []ChoiceValue{{ID: "encoding", Value: "utf8"}, {ID: "readonly", Value: "false"}}, CurrentFilter: 0}))

	r := wait(t, result)
	require.NoError(t, r.err)
	assert.Equal(t, ResponseOK, r.response)
	var uris []string
	require.NoError(t, r.results["uris"].Store(&uris))
	assert.Equal(t, []string{"file://" + filepath.ToSlash(folder) + "/a%20b.png"}, uris)
	var pairs []wireChoiceValue
	require.NoError(t, r.results["choices"].Store(&pairs))
	assert.Equal(t, []wireChoiceValue{{ID: "encoding", Value: "utf8"}, {ID: "readonly", Value: "false"}}, pairs)
	var current wireFilter
	require.NoError(t, r.results["current_filter"].Store(&current))
	assert.Equal(t, "Images", current.Name)
	assert.Equal(t, uint32(1), current.Patterns[0].Kind)
	assert.Empty(t, svc.Pending())
	assert.False(t, svc.Answer(req.ID, Answer{}))
}

func TestCloseCancelsTheDialog(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	start(t, address, ui)
	client := dbustest.Connect(t, address)

	result := call(client, "OpenFile", handle, "Open", map[string]dbus.Variant{})
	req := ui.next(t)
	require.NoError(t, client.Object(BusName, handle).Call(RequestInterface+".Close", 0).Err)

	assert.Equal(t, req.ID, ui.closedID(t))
	r := wait(t, result)
	require.NoError(t, r.err)
	assert.Equal(t, ResponseCancelled, r.response)
	assert.Empty(t, r.results)
	assert.Error(t, client.Object(BusName, handle).Call(RequestInterface+".Close", 0).Err)
}

func TestCallerVanishingCancels(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	svc := start(t, address, ui)
	caller, err := dbus.Connect(address)
	require.NoError(t, err)

	_ = call(caller, "SaveFile", handle, "Save", map[string]dbus.Variant{})
	req := ui.next(t)
	require.NoError(t, caller.Close())

	assert.Equal(t, req.ID, ui.closedID(t))
	assert.Eventually(t, func() bool { return len(svc.Pending()) == 0 }, 5*time.Second, 10*time.Millisecond)
}

func TestSaveFileNamesAndSingleURI(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	svc := start(t, address, ui)
	client := dbustest.Connect(t, address)
	folder := t.TempDir()

	result := call(client, "SaveFile", handle, "Save", map[string]dbus.Variant{
		"current_file": bytesOption(filepath.Join(folder, "notes.txt")),
	})
	req := ui.next(t)
	assert.Equal(t, MethodSave, req.Method)
	assert.Equal(t, folder, req.CurrentFolder)
	assert.Equal(t, "notes.txt", req.CurrentName)
	assert.Equal(t, filepath.Join(folder, "notes.txt"), req.CurrentFile)

	require.True(t, svc.Answer(req.ID, Answer{Paths: []string{filepath.Join(folder, "one.txt"), filepath.Join(folder, "two.txt")}}))
	r := wait(t, result)
	require.NoError(t, r.err)
	var uris []string
	require.NoError(t, r.results["uris"].Store(&uris))
	assert.Len(t, uris, 1)
	_, hasChoices := r.results["choices"]
	assert.False(t, hasChoices)
}

func TestSaveFilesCarriesNames(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	svc := start(t, address, ui)
	client := dbustest.Connect(t, address)
	folder := t.TempDir()

	result := call(client, "SaveFiles", handle, "Save", map[string]dbus.Variant{
		"current_folder": bytesOption(folder),
		"files":          dbus.MakeVariant([][]byte{[]byte("a.txt\x00"), []byte("b.txt\x00")}),
	})
	req := ui.next(t)
	assert.Equal(t, MethodSaveFiles, req.Method)
	assert.Equal(t, []string{"a.txt", "b.txt"}, req.Files)
	require.True(t, svc.Answer(req.ID, Answer{Paths: []string{filepath.Join(folder, "a.txt"), filepath.Join(folder, "b.txt")}}))
	r := wait(t, result)
	require.NoError(t, r.err)
	var uris []string
	require.NoError(t, r.results["uris"].Store(&uris))
	assert.Len(t, uris, 2)
}

func TestInvalidOptionsAreRefused(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	start(t, address, ui)
	client := dbustest.Connect(t, address)

	cases := map[string]struct {
		method  string
		h       dbus.ObjectPath
		options map[string]dbus.Variant
	}{
		"bad handle":       {"OpenFile", "/nope", nil},
		"filters type":     {"OpenFile", handle, map[string]dbus.Variant{"filters": dbus.MakeVariant("x")}},
		"filter kind":      {"OpenFile", handle, map[string]dbus.Variant{"filters": dbus.MakeVariant([]wireFilter{{Name: "x", Patterns: []wirePattern{{Kind: 2, Pattern: "*.a"}}}})}},
		"filter glob":      {"OpenFile", handle, map[string]dbus.Variant{"filters": dbus.MakeVariant([]wireFilter{{Name: "x", Patterns: []wirePattern{{Kind: 0, Pattern: "a[b"}}}})}},
		"filter mime":      {"OpenFile", handle, map[string]dbus.Variant{"filters": dbus.MakeVariant([]wireFilter{{Name: "x", Patterns: []wirePattern{{Kind: 1, Pattern: "image"}}}})}},
		"unknown current":  {"OpenFile", handle, map[string]dbus.Variant{"filters": dbus.MakeVariant([]wireFilter{{Name: "x", Patterns: []wirePattern{{Kind: 0, Pattern: "*.a"}}}}), "current_filter": dbus.MakeVariant(wireFilter{Name: "y", Patterns: []wirePattern{{Kind: 0, Pattern: "*.a"}}})}},
		"choice initial":   {"OpenFile", handle, map[string]dbus.Variant{"choices": dbus.MakeVariant([]wireChoice{{ID: "c", Label: "C", Options: []wireOption{{ID: "a", Label: "A"}}, Initial: "z"}})}},
		"choice boolean":   {"OpenFile", handle, map[string]dbus.Variant{"choices": dbus.MakeVariant([]wireChoice{{ID: "c", Label: "C", Options: []wireOption{}, Initial: "maybe"}})}},
		"multiple type":    {"OpenFile", handle, map[string]dbus.Variant{"multiple": dbus.MakeVariant("yes")}},
		"file name slash":  {"SaveFiles", handle, map[string]dbus.Variant{"files": dbus.MakeVariant([][]byte{[]byte("../x\x00")})}},
		"file name parent": {"SaveFiles", handle, map[string]dbus.Variant{"files": dbus.MakeVariant([][]byte{[]byte("..")})}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			options := tc.options
			if options == nil {
				options = map[string]dbus.Variant{}
			}
			r := wait(t, call(client, tc.method, tc.h, "x", options))
			require.Error(t, r.err)
			var dbusErr dbus.Error
			require.ErrorAs(t, r.err, &dbusErr)
			assert.Equal(t, errInvalidArgument, dbusErr.Name)
			select {
			case <-ui.shown:
				t.Fatal("the ui was shown for a refused request")
			default:
			}
		})
	}
}

func TestLenientOptionsAreDropped(t *testing.T) {
	address := dbustest.Bus(t)
	ui := newFakeUI()
	svc := start(t, address, ui)
	client := dbustest.Connect(t, address)

	result := call(client, "SaveFile", handle, "Save", map[string]dbus.Variant{
		"current_folder": bytesOption("/definitely/not/here"),
		"current_file":   bytesOption("relative/path.txt"),
		"current_name":   dbus.MakeVariant("bad/name"),
		"unknown":        dbus.MakeVariant(12),
	})
	req := ui.next(t)
	assert.Empty(t, req.CurrentFolder)
	assert.Empty(t, req.CurrentFile)
	assert.Empty(t, req.CurrentName)
	require.True(t, svc.Answer(req.ID, Answer{Response: ResponseCancelled}))
	r := wait(t, result)
	require.NoError(t, r.err)
	assert.Equal(t, ResponseCancelled, r.response)
}

func TestIntrospectionAndStatus(t *testing.T) {
	address := dbustest.Bus(t)
	svc := start(t, address, newFakeUI())
	client := dbustest.Connect(t, address)

	var xml string
	require.NoError(t, client.Object(BusName, Path).Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&xml))
	assert.Contains(t, xml, `name="SaveFiles"`)

	status := svc.SetEnabled(false)
	assert.False(t, status.Enabled)
	assert.False(t, status.Owned)
	assert.Error(t, client.Object(BusName, Path).Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Err)
	status = svc.SetEnabled(true)
	assert.True(t, status.Owned)
}

func TestStripMnemonic(t *testing.T) {
	assert.Equal(t, "Open", stripMnemonic("_Open"))
	assert.Equal(t, "Save_as", stripMnemonic("Save__as"))
	assert.Equal(t, "Use folder", stripMnemonic("Use fol_der"))
}
