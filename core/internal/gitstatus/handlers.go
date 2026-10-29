package gitstatus

import (
	"context"
	"errors"
	"strings"

	"github.com/duxlabs-dev/dux-go/files"
	"github.com/duxlabs-dev/dux-go/ipc"
	"github.com/duxlabs-dev/dux-go/ipc/params"
	"github.com/spalencsar/omarchy-files/core/internal/wire"
)

func Handles(method string) bool {
	return strings.HasPrefix(method, "git.")
}

func (s *Service) Handle(ctx context.Context, w *ipc.ConnWriter, req ipc.Request) {
	path := params.StringOpt(req.Params, "path", "")
	var (
		result Result
		err    error
	)
	switch req.Method {
	case "git.status":
		result, err = s.Status(ctx, path)
	case "git.stage":
		result, err = s.Stage(ctx, path, wire.StringList(req.Params, "paths"))
	case "git.unstage":
		result, err = s.Unstage(ctx, path, wire.StringList(req.Params, "paths"))
	case "git.stageAll":
		result, err = s.StageAll(ctx, path)
	case "git.unstageAll":
		result, err = s.UnstageAll(ctx, path)
	default:
		err = &files.Error{Code: files.CodeNotSupported, Err: errors.New("unknown Git method: " + req.Method)}
	}
	if err != nil {
		wire.RespondError(w, req.ID, err)
		return
	}
	w.Respond(req.ID, result)
}
