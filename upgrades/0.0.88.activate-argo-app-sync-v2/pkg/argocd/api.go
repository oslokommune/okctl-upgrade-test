package argocd

import (
	"fmt"
	"path"
)

// SetupApplicationsSync knows how to get ArgoCD to automatically synchronize a folder
func SetupApplicationsSync(opts SetupApplicationsSyncOpts) error {
	log := opts.Logger

	log.Info("Setting up application synchronization")

	relativeArgoCDManifestPath := path.Join(
		getArgoCDClusterConfigDir(opts.Cluster),
		defaultArgoCDSyncApplicationsManifestName,
	)
	relativeApplicationsSyncDir := path.Join(
		getArgoCDClusterConfigDir(opts.Cluster),
		defaultApplicationsSyncDirName,
	)

	log.Infof("Creating new application sync directory %s", relativeApplicationsSyncDir)

	err := createDirectory(opts.Fs, opts.DryRun, relativeApplicationsSyncDir)
	if err != nil {
		return fmt.Errorf("creating applications sync directory: %w", err)
	}

	log.Info("Installing ArgoCD application for application sync directory")

	err = installArgoCDApplicationForSyncDirectory(installArgoCDApplicationForSyncDirectoryOpts{
		DryRun:                        opts.DryRun,
		Fs:                            opts.Fs,
		Kubectl:                       opts.Kubectl,
		IACRepoURL:                    opts.Cluster.Github.URL(),
		ApplicationsSyncDir:           relativeApplicationsSyncDir,
		ArgoCDApplicationManifestPath: relativeArgoCDManifestPath,
	})
	if err != nil {
		return fmt.Errorf("installing ArgoCD application: %w", err)
	}

	return nil
}

// MigrateExistingApplicationManifests knows how to move all existing argocd-application manifests to the new sync
// directory
func MigrateExistingApplicationManifests(opts MigrateExistingApplicationManifestsOpts) error {
	log := opts.Logger

	rootAppDir := path.Join(opts.Cluster.Github.OutputPath, defaultApplicationsDirName)

	relativeArgoCDClusterConfigDir := getArgoCDClusterConfigDir(opts.Cluster)
	relativeAppSyncDir := path.Join(relativeArgoCDClusterConfigDir, defaultApplicationsSyncDirName)

	log.Infof("Migrating existing ArgoCD application manifests to %s\n", relativeAppSyncDir)

	argoCDApplicationManifestPaths, err := getAllArgoCDApplicationManifests(opts.Fs, rootAppDir)
	if err != nil {
		return fmt.Errorf("acquiring all ArgoCD application manifest paths: %w", err)
	}

	for _, sourcePath := range argoCDApplicationManifestPaths {
		appName := getApplicationNameFromPath(rootAppDir, sourcePath)

		destinationPath := path.Join(relativeAppSyncDir, fmt.Sprintf("%s.yaml", appName))

		log.Infof("Moving %s to %s\n", sourcePath, destinationPath)

		if opts.DryRun {
			continue
		}

		err = opts.Fs.Rename(sourcePath, destinationPath)
		if err != nil {
			return fmt.Errorf("moving file: %w", err)
		}
	}

	return nil
}
