package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/spalencsar/omarchy-files/core/internal/helper"
)

var helperCmd = &cobra.Command{
	Use:                "helper",
	Short:              "Root-side operation runner started by pkexec (not for direct use)",
	Hidden:             true,
	DisableFlagParsing: true,
	Args:               cobra.ArbitraryArgs,
	Run: func(_ *cobra.Command, args []string) {
		os.Exit(helper.Main(args, os.Stdin, os.Stdout, os.Stderr))
	},
}
