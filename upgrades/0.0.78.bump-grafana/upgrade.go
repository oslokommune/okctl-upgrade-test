package main

import "github.com/oslokommune/okctl-upgrade/upgrades/0.0.78.bump-grafana-v2/pkg/grafana"

func upgrade(context Context, flags cmdFlags) error {
	opts := grafana.Opts{
		DryRun:  flags.dryRun,
		Confirm: flags.confirm,
	}

	c := grafana.New(context.logger, opts)

	err := c.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
