package argocd

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/paths"
)

func EnableNamespacesSync(opts EnableNamespaceSyncOpts) error {
	absoluteNamespacesDir := path.Join(opts.AbsoluteRepositoryRootDirectory, paths.RelativeNamespacesDir(opts.Cluster))

	if !opts.DryRun {
		opts.Log.Debug("Preparing directory structure")

		err := opts.Fs.MkdirAll(absoluteNamespacesDir, paths.DefaultFolderPermissions)
		if err != nil {
			return fmt.Errorf("preparing namespaces dir: %w", err)
		}

		err = opts.Fs.WriteReader(
			path.Join(absoluteNamespacesDir, paths.DefaultReadmeFilename),
			strings.NewReader(namespacesReadmeTemplate),
		)
		if err != nil {
			return fmt.Errorf("creating namespaces readme: %w", err)
		}
	}

	opts.Log.Debug("Adding namespaces ArgoCD application")

	argoApp, err := scaffoldApplication(opts.Cluster, "namespaces", paths.RelativeNamespacesDir(opts.Cluster))
	if err != nil {
		return fmt.Errorf("scaffolding ArgoCD application: %w", err)
	}

	rawArgoApp, err := io.ReadAll(argoApp)
	if err != nil {
		return fmt.Errorf("buffering ArgoCD application: %w", err)
	}

	if !opts.DryRun {
		err = opts.Fs.WriteReader(
			path.Join(
				opts.AbsoluteRepositoryRootDirectory,
				paths.RelativeArgoCDConfigDir(opts.Cluster),
				"namespaces.yaml",
			),
			bytes.NewReader(rawArgoApp),
		)
		if err != nil {
			return fmt.Errorf("writing ArgoCD application: %w", err)
		}

		err = opts.Kubectl.Apply(bytes.NewReader(rawArgoApp))
		if err != nil {
			return fmt.Errorf("applying namespaces ArgoCD application: %w", err)
		}
	}

	return nil
}

//go:embed templates/namespaces-readme.md
var namespacesReadmeTemplate string
