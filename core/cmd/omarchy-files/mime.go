package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type mimeApp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type mimeQuery struct {
	Target    string    `json:"target"`
	MimeType  string    `json:"mimeType"`
	DefaultID string    `json:"defaultId"`
	Apps      []mimeApp `json:"apps"`
}

var mimeCmd = &cobra.Command{
	Use:   "mime",
	Short: "Look up mime types and default applications",
}

var mimeQueryCmd = &cobra.Command{
	Use:   "query <path|mime-type...>",
	Short: "Show the mime type, the default application and every handler",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return withSession(appServices, func(s *session) error {
			results := make([]mimeQuery, 0, len(args))
			for _, arg := range args {
				result, err := queryMime(s, arg)
				if err != nil {
					return err
				}
				results = append(results, result)
			}
			if jsonOutput {
				return printJSON(results)
			}
			for i, r := range results {
				if i > 0 {
					fmt.Println()
				}
				printMimeQuery(r, len(results) > 1)
			}
			return nil
		})
	},
}

var mimeDefaultCmd = &cobra.Command{
	Use:   "default <path|mime-type>",
	Short: "Print the default application's desktop id",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return withSession(appServices, func(s *session) error {
			result, err := queryMime(s, args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(map[string]string{"mimeType": result.MimeType, "defaultId": result.DefaultID})
			}
			if result.DefaultID == "" {
				return fmt.Errorf("no default application for %s", result.MimeType)
			}
			fmt.Println(result.DefaultID)
			return nil
		})
	},
}

var mimeSetCmd = &cobra.Command{
	Use:               "set <mime-type> <desktop-id>",
	Short:             "Set the default application for a mime type",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeDesktopIDs,
	RunE: func(_ *cobra.Command, args []string) error {
		desktopID := args[1]
		if !strings.HasSuffix(desktopID, ".desktop") {
			desktopID += ".desktop"
		}
		return withSession(appServices, func(s *session) error {
			var out map[string]string
			if err := s.call("files.apps.setDefault", map[string]any{"mime": args[0], "desktopId": desktopID}, &out); err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(out)
			}
			fmt.Printf("%s: %s\n", out["mimeType"], out["desktopId"])
			return nil
		})
	},
}

func queryMime(s *session, arg string) (mimeQuery, error) {
	params := map[string]any{"mime": arg}
	if looksLikePath(arg) {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return mimeQuery{}, err
		}
		params = map[string]any{"path": abs}
	}
	result := mimeQuery{Target: arg}
	if err := s.call("files.apps.forMime", params, &result); err != nil {
		return mimeQuery{}, err
	}
	return result, nil
}

func looksLikePath(arg string) bool {
	if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, ".") || !strings.Contains(arg, "/") {
		return true
	}
	_, err := os.Lstat(arg)
	return err == nil
}

func printMimeQuery(r mimeQuery, labelled bool) {
	if labelled {
		fmt.Printf("%s: %s\n", r.Target, r.MimeType)
	} else {
		fmt.Println(r.MimeType)
	}
	if r.DefaultID == "" {
		fmt.Println("default: none")
	} else {
		fmt.Printf("default: %s\n", r.DefaultID)
	}
	for _, app := range r.Apps {
		fmt.Printf("  %-40s %s\n", app.ID, app.Name)
	}
}

func completeDesktopIDs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var ids []string
	_ = withSession(appServices, func(s *session) error {
		var out struct {
			Apps []mimeApp `json:"apps"`
		}
		if err := s.call("files.apps.forMime", map[string]any{"mime": args[0]}, &out); err != nil {
			return err
		}
		for _, app := range out.Apps {
			ids = append(ids, app.ID+"\t"+app.Name)
		}
		return nil
	})
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	mimeCmd.AddCommand(mimeQueryCmd, mimeDefaultCmd, mimeSetCmd)
}
