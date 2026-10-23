package portalsetup

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const keyLine = Interface + "="

func TestPreferredLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		found   bool
	}{
		{"empty", "", "", false},
		{"no group", keyLine + "gtk;\n", "", false},
		{"other group", "[other]\n" + keyLine + "gtk;\n", "", false},
		{"commented", "[preferred]\n#" + keyLine + "gtk;\n", "", false},
		{"present", "[preferred]\ndefault=gtk;\n" + keyLine + "omarchy-files;\n", keyLine + "omarchy-files;", true},
		{"spaces", "[preferred]\n " + Interface + " = omarchy-files; \n", " " + Interface + " = omarchy-files; ", true},
		{"after other group", "[other]\n" + keyLine + "x;\n[preferred]\n" + keyLine + "y;\n", keyLine + "y;", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			line, found := preferredLine(tc.content)
			assert.Equal(t, tc.found, found)
			assert.Equal(t, tc.want, line)
		})
	}
}

func TestSetPreferred(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"empty", "", "[preferred]\n" + keyLine + "omarchy-files;\n"},
		{"no group", "# top\n[other]\nk=v\n", "# top\n[other]\nk=v\n\n[preferred]\n" + keyLine + "omarchy-files;\n"},
		{"no group no trailing newline", "[other]\nk=v", "[other]\nk=v\n\n[preferred]\n" + keyLine + "omarchy-files;\n"},
		{"group without key", "[preferred]\ndefault=gtk;\n\n[other]\nk=v\n", "[preferred]\ndefault=gtk;\n" + keyLine + "omarchy-files;\n\n[other]\nk=v\n"},
		{"empty group", "[preferred]\n", "[preferred]\n" + keyLine + "omarchy-files;\n"},
		{"replace", "[preferred]\n# c\n" + keyLine + "gtk;\ndefault=x;\n", "[preferred]\n# c\n" + keyLine + "omarchy-files;\ndefault=x;\n"},
		{"replace spaced key", "[preferred]\n" + Interface + " = gtk;\n", "[preferred]\n" + keyLine + "omarchy-files;\n"},
		{"key in other group untouched", "[other]\n" + keyLine + "gtk;\n", "[other]\n" + keyLine + "gtk;\n\n[preferred]\n" + keyLine + "omarchy-files;\n"},
		{"missing trailing newline", "[preferred]\n" + keyLine + "gtk;", "[preferred]\n" + keyLine + "omarchy-files;\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := setPreferred(tc.content, "omarchy-files;")
			assert.Equal(t, tc.want, got)
			assert.NotContains(t, got, "\r")
			assert.True(t, strings.HasSuffix(got, "\n"))
		})
	}
}

func TestRemovePreferred(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"empty", "", ""},
		{"absent", "[preferred]\ndefault=gtk;\n", "[preferred]\ndefault=gtk;\n"},
		{"other group untouched", "[other]\n" + keyLine + "gtk;\n", "[other]\n" + keyLine + "gtk;\n"},
		{"middle", "[preferred]\ndefault=gtk;\n" + keyLine + "omarchy-files;\n# tail\n", "[preferred]\ndefault=gtk;\n# tail\n"},
		{"only key", "[preferred]\n" + keyLine + "omarchy-files;\n", "[preferred]\n"},
		{"no trailing newline", "[preferred]\n" + keyLine + "omarchy-files;", "[preferred]\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, removePreferred(tc.content))
		})
	}
}

func TestRoundTripRestoresBytes(t *testing.T) {
	original := "# header\n[preferred]\n\tdefault = gtk;\n  " + Interface + "  =  gtk ; gnome ;  \n\n[other]\nx=1\n"
	line, found := preferredLine(original)
	require.True(t, found)
	changed := setPreferred(original, "omarchy-files;gtk;")
	assert.NotEqual(t, original, changed)
	assert.Equal(t, original, setPreferredLine(changed, line))
}

func TestBackendList(t *testing.T) {
	assert.Equal(t, []string{"gtk", "gnome"}, backendList(" gtk ; gnome ;"))
	assert.Nil(t, backendList(";;"))
	assert.Equal(t, []string{"*"}, backendList("*"))
}

func TestLookupKey(t *testing.T) {
	content := "[portal]\nDBusName = org.example\nInterfaces=a;b;\n[x]\nDBusName=other\n"
	v, ok := lookupKey(content, "portal", "DBusName")
	require.True(t, ok)
	assert.Equal(t, "org.example", v)
	v, ok = lookupKey(content, "x", "DBusName")
	require.True(t, ok)
	assert.Equal(t, "other", v)
	_, ok = lookupKey(content, "portal", "Missing")
	assert.False(t, ok)
	_, ok = lookupKey(content, "none", "DBusName")
	assert.False(t, ok)
}
