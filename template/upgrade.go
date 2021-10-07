package main

import (
	"github.com/oslokommune/okctl-upgrade/template/pkg/somecomponent"
)

func upgrade(context Context, flags cmdFlags) error {
	opts := somecomponent.Opts{
		DryRun:  flags.dryRun,
		Confirm: flags.confirm,
	}

	c := somecomponent.New(context.logger, opts)

	err := c.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
