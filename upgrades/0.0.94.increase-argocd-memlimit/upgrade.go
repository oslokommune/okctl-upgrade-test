package main

import (
	"fmt"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit/pkg/memlimit"
)

func upgrade(context Context, flags cmdflags.Flags) error {
	c, err := memlimit.New(context.logger, flags)
	if err != nil {
		return fmt.Errorf("creating increaser: %w", err)
	}

	err = c.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
