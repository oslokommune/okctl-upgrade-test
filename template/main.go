package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type cmdFlags struct {
	debug bool
	force bool
}

func buildRootCommand() *cobra.Command {
	flags := cmdFlags{}
	var context Context

	cmd := &cobra.Command{
		PreRunE: func(_ *cobra.Command, args []string) error {
			context = newContext(flags)
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if !flags.force {
				context.logger.Info("Simulating the upgrade, not doing any changes.")
			}

			err := upgrade(context)
			if err != nil {
				return fmt.Errorf("upgrade failed: %w", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&flags.debug, "debug", "d", false, "Set to true to enable debug output.")
	cmd.PersistentFlags().BoolVarP(&flags.force, "force", "f", false, "Set to true to apply changes. False means no changes are done.")

	return cmd
}
