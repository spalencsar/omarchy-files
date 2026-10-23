package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spalencsar/omarchy-files/core/internal/settings"
)

func TestWithTerminalDefault(t *testing.T) {
	kitty := Deps{Settings: func() settings.UISettings { return settings.UISettings{Terminal: "kitty"} }}
	none := Deps{Settings: func() settings.UISettings { return settings.UISettings{} }}

	tests := []struct {
		name   string
		deps   Deps
		params map[string]any
		want   any
	}{
		{"fills the setting", kitty, map[string]any{"path": "/tmp"}, "kitty"},
		{"fills into nil params", kitty, nil, "kitty"},
		{"keeps an explicit terminal", kitty, map[string]any{"terminal": "foot"}, "foot"},
		{"empty setting adds nothing", none, map[string]any{"path": "/tmp"}, nil},
		{"no settings source adds nothing", Deps{}, map[string]any{"path": "/tmp"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Method: "files.openTerminal", Params: tt.params}
			out := withTerminalDefault(req, tt.deps)
			assert.Equal(t, tt.want, out.Params["terminal"])
			if tt.params != nil && tt.params["terminal"] == nil {
				assert.NotContains(t, tt.params, "terminal", "the caller's map is not mutated")
			}
		})
	}
}
