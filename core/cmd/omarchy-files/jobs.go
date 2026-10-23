package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

const cliScope = "cli"

type jobView struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	State     string    `json:"state"`
	Sources   []string  `json:"sources"`
	Dest      string    `json:"dest,omitempty"`
	Skipped   int       `json:"skipped"`
	Cancelled bool      `json:"cancelled"`
	Error     string    `json:"error,omitempty"`
	Code      string    `json:"code,omitempty"`
	Conflict  *jobPause `json:"conflict,omitempty"`
	Failure   *jobPause `json:"failure,omitempty"`
	Password  *jobPause `json:"passwordPrompt,omitempty"`
}

type jobPause struct {
	ID      string   `json:"id"`
	Path    string   `json:"path"`
	Message string   `json:"message"`
	Actions []string `json:"actions"`
	Target  struct {
		Path string `json:"path"`
	} `json:"target"`
}

type jobEvent struct {
	Kind  string   `json:"kind"`
	JobID string   `json:"jobId"`
	Job   *jobView `json:"job"`
}

type jobResult struct {
	Job    jobView  `json:"job"`
	Failed []string `json:"failed"`
}

func runJobUnattended(s *session, method string, params map[string]any, onConflict string) (jobResult, error) {
	if err := s.subscribe("jobs"); err != nil {
		return jobResult{}, err
	}
	params["scope"] = cliScope
	var started struct {
		Job jobView `json:"job"`
	}
	if err := s.call(method, params, &started); err != nil {
		return jobResult{}, err
	}

	result := jobResult{Job: started.Job, Failed: []string{}}
	answered := ""
	for ev := range s.events {
		if ev.Topic != "jobs" {
			continue
		}
		var payload jobEvent
		if json.Unmarshal(ev.Data, &payload) != nil || payload.Kind != "job" || payload.JobID != started.Job.ID || payload.Job == nil {
			continue
		}
		job := *payload.Job
		result.Job = job
		switch job.State {
		case "done", "failed":
			return result, nil
		case "paused-conflict":
			if job.Conflict == nil || job.Conflict.ID == answered {
				continue
			}
			answered = job.Conflict.ID
			s.answer(job.ID, job.Conflict.ID, pick(job.Conflict.Actions, onConflict))
		case "paused-error":
			if job.Failure == nil || job.Failure.ID == answered {
				continue
			}
			answered = job.Failure.ID
			fmt.Fprintf(os.Stderr, "omarchy-files: %s: %s\n", job.Failure.Path, job.Failure.Message)
			result.Failed = append(result.Failed, job.Failure.Path)
			s.answer(job.ID, job.Failure.ID, pick(job.Failure.Actions, "skip"))
		case "paused-password":
			if job.Password == nil || job.Password.ID == answered {
				continue
			}
			answered = job.Password.ID
			fmt.Fprintf(os.Stderr, "omarchy-files: %s: needs a password\n", job.Password.Path)
			result.Failed = append(result.Failed, job.Password.Path)
			s.answer(job.ID, job.Password.ID, pick(job.Password.Actions, "skip"))
		}
	}
	return result, errors.New("lost the connection while the job ran")
}

func pick(actions []string, preferred string) string {
	if slices.Contains(actions, preferred) {
		return preferred
	}
	if slices.Contains(actions, "skip") {
		return "skip"
	}
	return "cancel"
}

func (s *session) answer(jobID, id, action string) {
	_ = s.call("files.op.respond", map[string]any{"jobId": jobID, "id": id, "action": action}, nil)
}

func (r jobResult) err() error {
	switch {
	case r.Job.State == "failed":
		return fmt.Errorf("%s failed: %s", r.Job.Kind, r.Job.Error)
	case r.Job.Cancelled:
		return fmt.Errorf("%s was cancelled", r.Job.Kind)
	case len(r.Failed) > 0:
		return fmt.Errorf("%s failed for %d item(s)", r.Job.Kind, len(r.Failed))
	}
	return nil
}

func (r jobResult) succeeded(source string) bool {
	for _, failed := range r.Failed {
		if failed == source || strings.HasPrefix(failed, source+"/") {
			return false
		}
	}
	return true
}

var conflictActions = map[string]string{"keep-both": "keepBoth", "skip": "skip", "replace": "replace"}

func conflictAction(flag string) (string, error) {
	action, ok := conflictActions[flag]
	if !ok {
		return "", fmt.Errorf("--conflict must be keep-both, skip or replace, not %q", flag)
	}
	return action, nil
}
