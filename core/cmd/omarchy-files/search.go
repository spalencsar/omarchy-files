package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type searchEvent struct {
	SearchID   string            `json:"searchId"`
	Kind       string            `json:"kind"`
	Entries    []json.RawMessage `json:"entries"`
	Count      int               `json:"count"`
	Truncated  bool              `json:"truncated"`
	Partial    bool              `json:"partial"`
	Cancelled  bool              `json:"cancelled"`
	Unreadable int               `json:"unreadable"`
	Backend    string            `json:"backend"`
	Phase      string            `json:"phase"`
	DepthLimit int               `json:"depthLimit"`
	Reason     string            `json:"reason"`
}

var (
	searchHiddenFlag      bool
	searchIgnoredFlag     bool
	searchFuzzyFlag       bool
	searchContentsFlag    bool
	searchDirectFlag      bool
	searchCrossDeviceFlag bool
	searchScopeFlag       string
	searchLimitFlag       int
	searchTypesFlag       []string
)

var searchCmd = &cobra.Command{
	Use:   "search <root> <query>",
	Short: "Search a folder tree by name, printing matching paths as they are found",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(_ *cobra.Command, args []string) error {
		root, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		params := map[string]any{"root": root, "query": args[1], "includeHidden": searchHiddenFlag, "includeIgnored": searchIgnoredFlag, "fuzzy": searchFuzzyFlag, "contents": searchContentsFlag, "direct": searchDirectFlag, "crossDevice": searchCrossDeviceFlag, "scope": searchScopeFlag}
		if searchLimitFlag > 0 {
			params["limit"] = searchLimitFlag
		}
		if len(searchTypesFlag) > 0 {
			params["types"] = searchTypesFlag
		}
		return withSession(searchServices, func(s *session) error {
			found := []json.RawMessage{}
			done, err := runSearch(s, params, func(raw json.RawMessage) {
				if jsonOutput {
					found = append(found, raw)
					return
				}
				var entry struct {
					Path string `json:"path"`
				}
				if json.Unmarshal(raw, &entry) == nil {
					fmt.Println(entry.Path)
				}
			})
			if err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(map[string]any{"entries": found, "count": done.Count, "truncated": done.Truncated, "partial": done.Partial, "unreadable": done.Unreadable, "backend": map[string]any{"name": done.Backend, "phase": done.Phase, "depthLimit": done.DepthLimit, "reason": done.Reason}})
			}
			if done.Backend == "index" {
				fmt.Fprintf(os.Stderr, "omarchy-files: served by the danksearch index (%s)\n", done.Phase)
			}
			if done.Partial {
				fmt.Fprintln(os.Stderr, "omarchy-files: some matches may be missing")
			}
			if done.Truncated {
				fmt.Fprintf(os.Stderr, "omarchy-files: stopped at %d results; raise --limit to see more\n", done.Count)
			}
			if done.Unreadable > 0 {
				fmt.Fprintf(os.Stderr, "omarchy-files: %d folders could not be read\n", done.Unreadable)
			}
			return nil
		})
	},
}

func runSearch(s *session, params map[string]any, onEntry func(json.RawMessage)) (searchEvent, error) {
	var started struct {
		SearchID string `json:"searchId"`
		Topic    string `json:"topic"`
	}
	if err := s.call("search.start", params, &started); err != nil {
		return searchEvent{}, err
	}
	if err := s.subscribe(started.Topic); err != nil {
		return searchEvent{}, err
	}

	var backend searchEvent
	for ev := range s.events {
		if ev.Topic != started.Topic {
			continue
		}
		var batch searchEvent
		if json.Unmarshal(ev.Data, &batch) != nil || batch.SearchID != started.SearchID {
			continue
		}
		switch batch.Kind {
		case "backend":
			backend = batch
		case "results":
			for _, raw := range batch.Entries {
				onEntry(raw)
			}
		case "done":
			batch.Backend, batch.Phase, batch.DepthLimit, batch.Reason = backend.Backend, backend.Phase, backend.DepthLimit, backend.Reason
			return batch, nil
		}
	}
	return searchEvent{}, errors.New("lost the connection while searching")
}

func init() {
	searchCmd.Flags().BoolVar(&searchHiddenFlag, "hidden", false, "Descend into and report hidden entries")
	searchCmd.Flags().BoolVar(&searchCrossDeviceFlag, "cross-device", false, "Follow into other filesystems")
	searchCmd.Flags().BoolVar(&searchIgnoredFlag, "include-ignored", false, "Walk heavy folders, gitignored entries and excluded folders too")
	searchCmd.Flags().BoolVar(&searchFuzzyFlag, "fuzzy", false, "Also accept terms whose letters appear in order")
	searchCmd.Flags().BoolVar(&searchContentsFlag, "contents", false, "Match indexed file contents (needs a running dsearch)")
	searchCmd.Flags().StringVar(&searchScopeFlag, "scope", "folder", "folder, home or everywhere")
	searchCmd.Flags().BoolVar(&searchDirectFlag, "direct", false, "Walk the disk even when the danksearch index could answer")
	searchCmd.Flags().IntVar(&searchLimitFlag, "limit", 0, "Stop after this many results (at most and by default 10000)")
	searchCmd.Flags().StringSliceVar(&searchTypesFlag, "type", nil, "Only documents, images, video, audio, archives or folders (repeatable)")
}
