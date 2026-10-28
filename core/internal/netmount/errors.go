package netmount

import (
	"errors"
	"fmt"
	"strings"

	"github.com/duxlabs-dev/dux-go/files"
	"github.com/godbus/dbus/v5"
	"github.com/spalencsar/omarchy-files/core/internal/dbusutil"
)

const CodeCanceled files.Code = "ECANCELED"

func fail(code files.Code, format string, args ...any) error {
	return &files.Error{Code: code, Err: fmt.Errorf(format, args...)}
}

// gvfs answers with unmapped GError names, so the message text decides.
func wrapDBus(err error) error {
	if err == nil {
		return nil
	}
	call, ok := err.(dbus.Error)
	if !ok {
		return &files.Error{Code: files.CodeIO, Err: err}
	}
	if call.Name == "org.freedesktop.DBus.Error.ServiceUnknown" {
		return fail(files.CodeNotSupported, "gvfs is not running")
	}
	text := dbusutil.ErrorMessage(call)
	return &files.Error{Code: codeForText(call.Name, text), Err: errors.New(text)}
}

func codeForText(name, text string) files.Code {
	switch {
	case strings.Contains(name, "PermissionDenied"), strings.Contains(text, "Permission denied"):
		return files.CodeAccess
	case strings.Contains(text, "cancelled"):
		return CodeCanceled
	case strings.Contains(text, "already mounted"):
		return files.CodeExists
	case strings.Contains(text, "not mounted"):
		return files.CodeInvalid
	case strings.Contains(text, "busy"):
		return files.CodeBusy
	}
	return files.CodeIO
}
