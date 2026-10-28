package devices

import (
	"errors"
	"fmt"
	"strings"

	"github.com/duxlabs-dev/dux-go/files"
	"github.com/godbus/dbus/v5"
	"github.com/spalencsar/omarchy-files/core/internal/dbusutil"
)

const CodePerm files.Code = "EPERM"

func fail(code files.Code, format string, args ...any) error {
	return &files.Error{Code: code, Err: fmt.Errorf(format, args...)}
}

func wrapDBus(err error) error {
	if err == nil {
		return nil
	}
	call, ok := err.(dbus.Error)
	if !ok {
		return &files.Error{Code: files.CodeIO, Err: err}
	}
	return &files.Error{Code: codeForName(call.Name), Err: errors.New(dbusutil.ErrorMessage(call))}
}

func codeForName(name string) files.Code {
	switch strings.TrimPrefix(name, BusName+".Error.") {
	case "NotAuthorized", "NotAuthorizedCanObtain", "NotAuthorizedDismissed":
		return files.CodeAccess
	case "DeviceBusy":
		return files.CodeBusy
	case "AlreadyMounted", "AlreadyUnmounting":
		return files.CodeExists
	case "NotMounted", "OptionNotPermitted", "MountedByOtherUser":
		return files.CodeInvalid
	case "NotSupported":
		return files.CodeNotSupported
	}
	switch name {
	case "org.freedesktop.DBus.Error.ServiceUnknown", "org.freedesktop.DBus.Error.NameHasNoOwner":
		return files.CodeNotSupported
	case "org.freedesktop.DBus.Error.AccessDenied":
		return files.CodeAccess
	case "org.freedesktop.DBus.Error.UnknownMethod", "org.freedesktop.DBus.Error.UnknownInterface":
		return files.CodeNotSupported
	case "org.freedesktop.DBus.Error.UnknownObject":
		return files.CodeNotFound
	}
	return files.CodeIO
}
