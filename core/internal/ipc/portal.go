package ipc

import (
	"context"
	"encoding/json"

	"github.com/spalencsar/omarchy-files/core/internal/filechooser"
	"github.com/duxlabs-dev/dux-go/ipc/params"
)

type ChooserUI struct{ Deps Deps }

func (u ChooserUI) ShowChooser(req filechooser.Request) {
	publishUI(u.Deps, map[string]any{"action": "chooser", "request": req})
}

func (u ChooserUI) CloseChooser(id string) {
	publishUI(u.Deps, map[string]any{"action": "chooserClose", "id": id})
}

func republishChoosers(deps Deps) {
	if deps.Portal == nil {
		return
	}
	for _, req := range deps.Portal.Pending() {
		deps.Bus.Publish("ui", map[string]any{"action": "chooser", "request": req})
	}
}

func HandlePortal(_ context.Context, w *ConnWriter, req Request, deps Deps) {
	switch req.Method {
	case "portal.respond":
		if deps.Portal == nil {
			w.RespondError(req.ID, "portal backend unavailable")
			return
		}
		id := params.StringOpt(req.Params, "id", "")
		answer, err := answerFrom(req.Params)
		switch {
		case id == "":
			w.RespondError(req.ID, "id is required")
		case err != nil:
			w.RespondError(req.ID, "portal.respond: "+err.Error())
		case !deps.Portal.Answer(id, answer):
			w.RespondError(req.ID, "no open request "+id)
		default:
			w.Respond(req.ID, map[string]any{"ok": true})
		}
	default:
		w.RespondError(req.ID, "unknown portal method: "+req.Method)
	}
}

func answerFrom(params map[string]any) (filechooser.Answer, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return filechooser.Answer{}, err
	}
	answer := filechooser.Answer{CurrentFilter: -1}
	if err := json.Unmarshal(data, &answer); err != nil {
		return filechooser.Answer{}, err
	}
	if answer.Paths == nil {
		answer.Paths = []string{}
	}
	return answer, nil
}
