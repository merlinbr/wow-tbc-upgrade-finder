package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

type rankUpgradesOptions struct {
	addr      string
	noBrowser bool
}

func newRankUpgradesCommand(version string) *cobra.Command {
	opts := &rankUpgradesOptions{}
	cmd := &cobra.Command{
		Use:   "rank-upgrades",
		Short: "rank practical single-item DPS upgrades",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			server := newUpgradeServer(version)
			return server.Serve(ctx, opts.addr, !opts.noBrowser)
		},
	}

	cmd.Flags().StringVar(&opts.addr, "addr", "127.0.0.1:0", "address to bind server to")
	cmd.Flags().BoolVar(&opts.noBrowser, "no-browser", false, "do not open browser automatically")

	return cmd
}
