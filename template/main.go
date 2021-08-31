package main

import (
	"fmt"
	"github.com/oslokommune/okctl-upgrade/template/pkg/logger"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type Flags struct {
	debug bool
	force bool
}

type Context struct {
	logger logger.Logger
	force  bool
}

func buildRootCommand() *cobra.Command {
	flags := Flags{}
	var context Context

	cmd := &cobra.Command{
		SilenceUsage: true,
		PreRunE: func(_ *cobra.Command, args []string) error {
			context = newContext(flags)
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
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

func newContext(flags Flags) Context {
	var level logger.Level
	if flags.debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		logger: logger.New(level),
		force:  flags.force,
	}
}
