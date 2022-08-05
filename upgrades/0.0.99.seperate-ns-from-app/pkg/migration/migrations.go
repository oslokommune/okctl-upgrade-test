package migration

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/paths"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
)

func migrateApplication(logger debugLogger, dryRun bool, fs *afero.Afero, cluster v1alpha1.Cluster, absoluteRepositoryRoot string, appName string) error {
	absoluteNamespacesDir := path.Join(absoluteRepositoryRoot, paths.RelativeNamespacesDir(cluster))
	absoluteApplicationBaseDir := path.Join(
		absoluteRepositoryRoot,
		cluster.Github.OutputPath,
		paths.ApplicationsDir,
		appName,
		paths.ApplicationBaseDir,
	)

	logger.Debug(fmt.Sprintf("Migrating %s", appName))

	sourcePath := path.Join(absoluteApplicationBaseDir, "namespace.yaml")

	namespaceName, err := getNamespaceName(fs, sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debug("Namespace owned by application not found, ignoring")

			return nil
		}

		return fmt.Errorf("acquiring namespace name: %w", err)
	}

	logger.Debug(fmt.Sprintf("Namespace %s owned by application found, migrating", namespaceName))

	destinationPath := path.Join(absoluteNamespacesDir, fmt.Sprintf("%s.yaml", namespaceName))

	if dryRun {
		return nil
	}

	err = copyFile(fs, sourcePath, destinationPath)
	if err != nil {
		return fmt.Errorf("copying: %w", err)
	}

	return nil
}

func removeRedundantNamespacesFromBase(logger debugLogger, dryRun bool, fs *afero.Afero, cluster v1alpha1.Cluster, absoluteRepositoryRoot string) error {
	apps, err := getApplicationsInCluster(fs, cluster, absoluteRepositoryRoot)
	if err != nil {
		return fmt.Errorf("acquiring apps: %w", err)
	}

	logger.Debug("Removing redundant application owned namespaces")

	for _, app := range apps {
		logger.Debug(fmt.Sprintf("Checking %s for redundant namespaces", app))

		migrated, err := isFullyMigrated(fs, cluster, absoluteRepositoryRoot, app)
		if err != nil {
			return fmt.Errorf("checking for base namespace: %w", err)
		}

		if migrated {
			logger.Debug("Fully migrated, ignoring")

			continue
		}

		absoluteApplicationDir := path.Join(absoluteRepositoryRoot, cluster.Github.OutputPath, paths.ApplicationsDir, app)
		absoluteNamespacePath := path.Join(absoluteApplicationDir, paths.ApplicationBaseDir, "namespace.yaml")

		cleanable, err := isCleanable(fs, absoluteRepositoryRoot, cluster, app)
		if err != nil {
			return fmt.Errorf("checking if all adjacent clusters has namespace: %w", err)
		}

		if cleanable {
			logger.Debug("Found redundant namespace, removing")

			if dryRun {
				continue
			}

			err = fs.Remove(absoluteNamespacePath)
			if err != nil {
				return fmt.Errorf("removing base namespace: %w", err)
			}

			err = deleteNamespaceEntryFromKustomizationResources(fs, path.Join(
				absoluteApplicationDir,
				paths.ApplicationBaseDir,
				paths.DefaultKustomizationFilename,
			))
			if err != nil {
				return fmt.Errorf("deleting namespace entry from kustomization: %w", err)
			}
		}
	}

	return nil
}

func isFullyMigrated(fs *afero.Afero, cluster v1alpha1.Cluster, absoluteRepositoryRoot string, appName string) (bool, error) {
	absAppBasePath := path.Join(
		absoluteRepositoryRoot,
		cluster.Github.OutputPath,
		paths.ApplicationsDir,
		appName,
		paths.ApplicationBaseDir,
	)

	exists, err := fs.Exists(path.Join(absAppBasePath, "namespace.yaml"))
	if err != nil {
		return false, fmt.Errorf(": %w", err)
	}

	return !exists, nil
}

func clusterHasNewStyleNamespace(fs *afero.Afero, absoluteRepositoryOutputDir string, clusterName string, namespaceName string) (bool, error) {
	potentialNamespacePath := path.Join(
		absoluteRepositoryOutputDir,
		clusterName,
		paths.ArgocdConfigDir,
		"namespaces",
		fmt.Sprintf("%s.yaml", namespaceName),
	)

	exists, err := fs.Exists(potentialNamespacePath)
	if err != nil {
		return false, fmt.Errorf("checking existence: %w", err)
	}

	return exists, nil
}

func getAssociatedClusters(fs *afero.Afero, absoluteApplicationDir string) ([]string, error) {
	items, err := fs.ReadDir(path.Join(absoluteApplicationDir, paths.ApplicationOverlaysDir))
	if err != nil {
		return nil, fmt.Errorf("listing directory: %w", err)
	}

	relevantClusters := make([]string, 0)

	for _, item := range items {
		if item.IsDir() {
			relevantClusters = append(relevantClusters, item.Name())
		}
	}

	return relevantClusters, nil
}

// An app is cleanable IF and only IF all associated clusters has a new style namespace for the select namespace
func isCleanable(fs *afero.Afero, absoluteRepositoryRoot string, cluster v1alpha1.Cluster, appName string) (bool, error) {
	absoluteApplicationDir := path.Join(absoluteRepositoryRoot, paths.RelativeApplicationDir(cluster, appName))

	relevantClusters, err := getAssociatedClusters(fs, absoluteApplicationDir)
	if err != nil {
		return false, fmt.Errorf("acquiring relevant clusters: %w", err)
	}

	namespaceName, err := getNamespaceName(
		fs,
		path.Join(absoluteApplicationDir, paths.ApplicationBaseDir, "namespace.yaml"),
	)
	if err != nil {
		return false, fmt.Errorf("acquiring namespace name: %w", err)
	}

	for _, clusterName := range relevantClusters {
		hasNewStyleNamespace, err := clusterHasNewStyleNamespace(
			fs,
			path.Join(absoluteRepositoryRoot, cluster.Github.OutputPath),
			clusterName,
			namespaceName,
		)
		if err != nil {
			return false, fmt.Errorf("checking for namespace: %w", err)
		}

		if !hasNewStyleNamespace {
			return false, nil
		}
	}

	return true, nil
}
