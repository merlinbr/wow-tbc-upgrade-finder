package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "wowsimcli",
		Short: "wowsims command line tool",
		Long:  "wowsims command line tool",
	}

	rootCmd.AddCommand(newVersionCommand(version))
	rootCmd.AddCommand(simCmd)
	rootCmd.AddCommand(decodeLinkCmd)
	rootCmd.AddCommand(newRankUpgradesCommand(version))

	return rootCmd
}

func Execute(version string) {
	rootCmd := newRootCommand(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
