package argocd

import (
	"bytes"
	"fmt"

	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
	"github.com/oslokommune/okctl/pkg/scaffold"
	"github.com/oslokommune/okctl/pkg/scaffold/resources"
	"github.com/spf13/afero"
)

func createDirectory(fs *afero.Afero, dryRun bool, path string) error {
	if dryRun {
		return nil
	}

	err := fs.MkdirAll(path, defaultFolderPermissions)
	if err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	return nil
}

func installArgoCDApplicationForSyncDirectory(opts installArgoCDApplicationForSyncDirectoryOpts) error {
	app := v1alpha1.Application{Metadata: v1alpha1.ApplicationMeta{
		Name:      defaultArgoCDSyncApplicationName,
		Namespace: defaultArgoCDSyncApplicationNamespace,
	}}

	argoCDApplication := resources.CreateArgoApp(app, opts.IACRepoURL, opts.ApplicationsSyncDir)
	argoCDApplication.Spec.SyncPolicy.Automated.Prune = true

	rawArgoCDApplication, err := scaffold.ResourceAsBytes(argoCDApplication)
	if err != nil {
		return fmt.Errorf("marshalling ArgoCD application manifest: %w", err)
	}

	if opts.DryRun {
		return nil
	}

	err = opts.Fs.WriteReader(opts.ArgoCDApplicationManifestPath, bytes.NewReader(rawArgoCDApplication))
	if err != nil {
		return fmt.Errorf("writing ArgoCD application manifest: %w", err)
	}

	err = opts.Kubectl.Apply(bytes.NewReader(rawArgoCDApplication))
	if err != nil {
		return fmt.Errorf("applying ArgoCD application manifest: %w", err)
	}

	return nil
}
