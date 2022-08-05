package migration

import (
	"fmt"
)

func MigrateExistingApplicationNamespacesToCluster(opts MigrateExistingApplicationNamespacesToClusterOpts) error {
	apps, err := getApplicationsInCluster(opts.Fs, opts.Cluster, opts.AbsoluteRepositoryRoot)
	if err != nil {
		return fmt.Errorf("scanning for applications: %w", err)
	}

	for _, app := range apps {
		err = migrateApplication(opts.Log, opts.DryRun, opts.Fs, opts.Cluster, opts.AbsoluteRepositoryRoot, app)
		if err != nil {
			return fmt.Errorf("migrating %s: %w", app, err)
		}
	}

	opts.Log.Debug("Cleaning up redundant application owned namespaces")

	err = removeRedundantNamespacesFromBase(opts.Log, opts.DryRun, opts.Fs, opts.Cluster, opts.AbsoluteRepositoryRoot)
	if err != nil {
		return fmt.Errorf("removing redundant namespace manifests from application base folders: %w", err)
	}

	return nil
}
