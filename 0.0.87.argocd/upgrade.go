package main

import (
	"fmt"
	argocdPkg "github.com/oslokommune/okctl-upgrade/0.0.87.argocd/pkg/argocd"
	"github.com/oslokommune/okctl-upgrade/0.0.87.argocd/pkg/lib/cmdflags"
)

func upgrade(context Context, flags cmdflags.Flags) error {
	argocd, err := argocdPkg.New(context.log, flags)
	if err != nil {
		return fmt.Errorf("creating argocd: %w", err)
	}

	err = argocd.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
