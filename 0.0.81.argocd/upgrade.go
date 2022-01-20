package main

import (
	"fmt"
	argocdPkg "github.com/oslokommune/okctl-upgrade/0.0.81.argocd/pkg/argocd"
)

func upgrade(context Context, flags cmdFlags) error {
	opts := argocdPkg.Opts{
		DryRun:        flags.dryRun,
		Confirm:       flags.confirm,
		SkipPreflight: flags.skipPreflight,
	}

	argocd, err := argocdPkg.New(context.log, opts)
	if err != nil {
		return fmt.Errorf("creating argocd: %w", err)
	}

	err = argocd.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
