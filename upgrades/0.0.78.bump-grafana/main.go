package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.78.bump-grafana-v2/pkg/commonerrors"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()

	err := cmd.Execute()

	if err != nil && errors.Is(err, commonerrors.ErrUserAborted) {
		fmt.Println("Upgrade aborted by user.")
	} else if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	if err != nil {
		os.Exit(1)
	}
}

type cmdFlags struct {
	debug   bool
	dryRun  bool
	confirm bool
}

func buildRootCommand() *cobra.Command {
	flags := cmdFlags{}

	var context Context

	filename := filepath.Base(os.Args[0])

	cmd := &cobra.Command{
		Short:         "Upgrades an okctl cluster",
		Long:          "Note, boolean arguments must be specified on the form --arg=bool (and not on the form --arg bool).",
		Use:           filename,
		Example:       fmt.Sprintf("%s --debug=false", filename),
		SilenceErrors: true, // true as we print errors in the main() function
		SilenceUsage:  true, // true because we don't want to show usage if an errors occurs
		PreRunE: func(_ *cobra.Command, args []string) error {
			context = newContext(flags)
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return upgrade(context, flags)
		},
	}

	/*
	 * Flags supported. Expected behavior is as following:
	 *
	 * --debug:		Outputs extra output for debugging.
	 *
	 * --dry-run: 	If set to true, the upgrade will not make any changes, but only print what would be done, as if
	 * 				running a simulation.
	 *				If set to false, the upgrade will make actual changes.
	 *
	 * --confirm:	Skips all confirmation prompts, if any.
	 */
	cmd.PersistentFlags().BoolVarP(&flags.debug,
		"debug", "d", false, "Set this to enable debug output.")
	cmd.PersistentFlags().BoolVarP(&flags.dryRun,
		"dry-run", "n", true, "Don't actually do any changes, just show what would be done.")
	cmd.PersistentFlags().BoolVarP(&flags.confirm,
		"confirm", "c", false, "Set this to skip confirmation prompts.")

	return cmd
}
