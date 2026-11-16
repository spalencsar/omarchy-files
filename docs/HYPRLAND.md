# Optional Hyprland integration

Omarchy Files does not install Hyprland bindings or window rules. Copy the
snippets below into your user config under `~/.config/hypr/` if you want them.
Do not put them under `/usr/share/omarchy`.

The stock Omarchy file-manager keys are `SUPER + SHIFT + F` (open) and
`SUPER + ALT + SHIFT + F` (open the focused terminal's directory). Those launch
Nautilus. Rebind them only if you want Omarchy Files on the same keys.

Window class: `io.github.spalencsar.OmarchyFiles`.

## Window rules

Save as `~/.config/hypr/omarchy-files.lua` and add
`require("hypr.omarchy-files")` at the end of `~/.config/hypr/hyprland.lua`.

```lua
-- Slight transparency, same treatment as Nautilus.
o.transparent_window("io.github.spalencsar.OmarchyFiles")

-- Open/save chooser windows share the app id; float them like other pickers.
-- Titles follow the UI language (English and German defaults shown here).
o.window({
  class = "io.github.spalencsar.OmarchyFiles",
  title = "^(Open file|Open files|Save file|Save files|Select folder|Datei öffnen|Dateien öffnen|Datei speichern|Dateien speichern|Ordner wählen|Open.*Files?|Open [Ff]older.*|Save.*As|Save|All Files|.*wants to [open|save].*|[Cc]hoose.*)",
}, { tag = "+floating-window" })
```

Reload and check:

```bash
hyprctl reload
hyprctl configerrors
```

## Keybindings

Add to `~/.config/hypr/bindings.lua`. `o.rebind` replaces the stock Nautilus
bindings on those keys.

```lua
o.rebind("SUPER + SHIFT + F", "File manager", {
  launch = "omarchy-files show",
  focus = "io.github.spalencsar.OmarchyFiles",
})
o.rebind(
  "SUPER + ALT + SHIFT + F",
  "File manager (cwd)",
  'uwsm-app -- omarchy-files open "$(omarchy-cmd-terminal-cwd)"'
)
```

To keep Nautilus on those keys, leave `bindings.lua` unchanged and launch
Omarchy Files from the bar plugin, the desktop entry, or `omarchy-files show`.
