package upgrade

import (
	"context"
	"fmt"
	"path"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/commonerrors"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/kubectl"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/argocd"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/migration"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/paths"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/logging"
)

func Start(_ context.Context, logger logging.Logger, fs *afero.Afero, flags cmdflags.Flags, cluster v1alpha1.Cluster) error {
	absoluteRepositoryRootDir, err := paths.GetRepositoryRootDirectory()
	if err != nil {
		return fmt.Errorf("acquiring repository root dir: %w", err)
	}

	logger.Debug("Running preflight tests")

	err = preflight(logger, fs, absoluteRepositoryRootDir, cluster)
	if err != nil {
		return fmt.Errorf("running preflight tests: %w", err)
	}

	logger.Debug("Enabling namespace synchronization")

	err = argocd.EnableNamespacesSync(argocd.EnableNamespaceSyncOpts{
		Log:                             &logger,
		DryRun:                          flags.DryRun,
		Fs:                              fs,
		Kubectl:                         &kubectl.Client{},
		AbsoluteRepositoryRootDirectory: absoluteRepositoryRootDir,
		Cluster:                         cluster,
	})
	if err != nil {
		return fmt.Errorf("adding namespaces app manifest: %w", err)
	}

	logger.Debug("Migrate application owned namespaces")

	err = migration.MigrateExistingApplicationNamespacesToCluster(migration.MigrateExistingApplicationNamespacesToClusterOpts{
		Log:                    &logger,
		DryRun:                 flags.DryRun,
		Fs:                     fs,
		Cluster:                cluster,
		AbsoluteRepositoryRoot: absoluteRepositoryRootDir,
	})
	if err != nil {
		return fmt.Errorf("migrating existing namespaces: %w", err)
	}

	return nil
}

func preflight(logger logging.Logger, fs *afero.Afero, absoluteRepositoryRootDir string, cluster v1alpha1.Cluster) error {
	absoluteApplicationsDir := path.Join(absoluteRepositoryRootDir, cluster.Github.OutputPath, paths.ApplicationsDir)

	hasApplicationsDir, err := fs.Exists(absoluteApplicationsDir)
	if err != nil {
		return fmt.Errorf("checking applications directory existence: %w", err)
	}

	if !hasApplicationsDir {
		logger.Debug("No application directory found, ignoring upgrade")

		return fmt.Errorf("no application directory found: %w", commonerrors.ErrNothingToDo)
	}

	return nil
}
