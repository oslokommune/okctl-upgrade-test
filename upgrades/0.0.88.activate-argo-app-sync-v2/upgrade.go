package main

import (
	"fmt"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v2/pkg/argocd"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v2/pkg/kubectl/binary"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v2/pkg/okctl"
)

func upgrade(upgradeContext Context, flags cmdFlags) error {
	o, err := okctl.InitializeOkctl()
	if err != nil {
		return fmt.Errorf("initializing okctl: %w", err)
	}

	kubectlClient := binary.New(binary.NewOpts{
		Logger:              upgradeContext.logger,
		Fs:                  upgradeContext.Fs,
		BinaryProvider:      o.BinariesProvider,
		CredentialsProvider: o.CredentialsProvider,
		Cluster:             *o.Declaration,
	})

	err = argocd.SetupApplicationsSync(argocd.SetupApplicationsSyncOpts{
		Logger:  upgradeContext.logger,
		Fs:      upgradeContext.Fs,
		Cluster: *o.Declaration,
		Kubectl: kubectlClient,
		DryRun:  flags.dryRun,
	})
	if err != nil {
		return fmt.Errorf("activating application folder synchronization: %w", err)
	}

	upgradeContext.logger.Info("Migrating existing application manifests to new location")

	err = argocd.MigrateExistingApplicationManifests(argocd.MigrateExistingApplicationManifestsOpts{
		Logger:  upgradeContext.logger,
		DryRun:  flags.dryRun,
		Fs:      upgradeContext.Fs,
		Cluster: *o.Declaration,
	})
	if err != nil {
		return fmt.Errorf("migrating existing ArgoCD application manifests: %w", err)
	}

	return nil
}
