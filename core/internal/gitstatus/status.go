package gitstatus

import (
	"path/filepath"
	"sort"
	"strings"
)

type state struct {
	code     string
	staged   bool
	unstaged bool
}

var priorities = map[string]int{
	"modified":  1,
	"renamed":   2,
	"added":     3,
	"untracked": 4,
	"deleted":   5,
	"conflict":  6,
}

func parseStatus(root string, data []byte) Result {
	result := Result{}
	states := map[string]state{}
	records := strings.Split(string(data), "\x00")
	for i := 0; i < len(records); i++ {
		record := records[i]
		if record == "" {
			continue
		}
		if strings.HasPrefix(record, "# branch.head ") {
			result.Branch = strings.TrimPrefix(record, "# branch.head ")
			result.Detached = result.Branch == "(detached)"
			continue
		}
		path, current, renamed := parseRecord(record)
		if path == "" || current.code == "" {
			continue
		}
		if renamed && i+1 < len(records) {
			i++
		}
		states[filepath.Clean(filepath.Join(root, filepath.FromSlash(path)))] = current
	}
	aggregates := map[string]state{}
	for path, current := range states {
		result.Changed++
		if current.staged {
			result.Staged++
		}
		if current.unstaged {
			result.Unstaged++
		}
		result.Entries = append(result.Entries, entryFor(path, current, false))
		for parent := filepath.Dir(path); parent != root && within(root, parent); parent = filepath.Dir(parent) {
			aggregates[parent] = mergeState(aggregates[parent], current)
		}
	}
	for path, current := range aggregates {
		if _, exists := states[path]; exists {
			continue
		}
		result.Entries = append(result.Entries, entryFor(path, current, true))
	}
	sort.Slice(result.Entries, func(i, j int) bool { return result.Entries[i].Path < result.Entries[j].Path })
	return result
}

func mergeState(previous, current state) state {
	merged := previous
	if merged.code == "" || priorities[current.code] > priorities[merged.code] {
		merged.code = current.code
	}
	merged.staged = previous.staged || current.staged
	merged.unstaged = previous.unstaged || current.unstaged
	return merged
}

func parseRecord(record string) (string, state, bool) {
	switch record[0] {
	case '?':
		return strings.TrimPrefix(record, "? "), state{code: "untracked", unstaged: true}, false
	case 'u':
		fields := strings.SplitN(record, " ", 11)
		if len(fields) < 11 {
			return "", state{}, false
		}
		return fields[10], state{code: "conflict", staged: true, unstaged: true}, false
	case '1':
		fields := strings.SplitN(record, " ", 9)
		if len(fields) < 9 {
			return "", state{}, false
		}
		return fields[8], stateFor(fields[1]), false
	case '2':
		fields := strings.SplitN(record, " ", 10)
		if len(fields) < 10 {
			return "", state{}, false
		}
		current := stateFor(fields[1])
		if current.code == "modified" {
			current.code = "renamed"
		}
		return fields[9], current, true
	default:
		return "", state{}, false
	}
}

func stateFor(xy string) state {
	if len(xy) < 2 {
		return state{}
	}
	x, y := xy[0], xy[1]
	current := state{staged: x != '.', unstaged: y != '.'}
	switch {
	case strings.ContainsRune(xy, 'U') || xy == "AA" || xy == "DD":
		current.code = "conflict"
	case x == 'D' || y == 'D':
		current.code = "deleted"
	case x == 'A':
		current.code = "added"
	case x == 'R' || x == 'C':
		current.code = "renamed"
	default:
		current.code = "modified"
	}
	return current
}

func entryFor(path string, current state, aggregate bool) Entry {
	return Entry{Path: path, Code: current.code, Label: labelFor(current.code), Staged: current.staged, Unstaged: current.unstaged, Aggregate: aggregate}
}

func labelFor(code string) string {
	switch code {
	case "added":
		return "Added"
	case "untracked":
		return "Untracked"
	case "deleted":
		return "Deleted"
	case "renamed":
		return "Renamed"
	case "conflict":
		return "Conflict"
	default:
		return "Modified"
	}
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
