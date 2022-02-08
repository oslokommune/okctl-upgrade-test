package main

import (
	"fmt"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/rotatesshkey"
)

func upgrade(context Context, flags cmdflags.Flags) error {
	c, err := rotatesshkey.New(context.logger, flags)
	if err != nil {
		return fmt.Errorf("creating rotater: %w", err)
	}

	err = c.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
