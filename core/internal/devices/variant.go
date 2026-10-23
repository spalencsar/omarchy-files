package devices

import (
	"github.com/spalencsar/omarchy-files/core/internal/dbusutil"
	"github.com/godbus/dbus/v5"
)

func stringOf(v dbus.Variant) string {
	s, _ := v.Value().(string)
	return s
}

func boolOf(v dbus.Variant) bool {
	b, _ := v.Value().(bool)
	return b
}

func uint64Of(v dbus.Variant) uint64 {
	n, _ := v.Value().(uint64)
	return n
}

func uint32Of(v dbus.Variant) uint32 {
	n, _ := v.Value().(uint32)
	return n
}

func objectOf(v dbus.Variant) string {
	path, _ := v.Value().(dbus.ObjectPath)
	if path == "" || path == "/" {
		return ""
	}
	return string(path)
}

func byteStringOf(v dbus.Variant) string {
	raw, _ := v.Value().([]byte)
	return dbusutil.ByteString(raw)
}
