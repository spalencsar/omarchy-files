package main

import (
	"encoding/json"
	"os"
)

var jsonOutput bool

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
