package migration

import (
	"fmt"
	"path"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/argocd"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/paths"

	"sigs.k8s.io/yaml"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
)

func getApplicationsInCluster(fs *afero.Afero, cluster v1alpha1.Cluster, absoluteRepositoryRoot string) ([]string, error) {
	absoluteApplicationsDir := path.Join(absoluteRepositoryRoot, paths.RelativeArgoCDApplicationsDir(cluster))

	files, err := fs.ReadDir(absoluteApplicationsDir)
	if err != nil {
		return nil, fmt.Errorf("retrieving items in app dir: %w", err)
	}

	apps := make([]string, 0)

	for _, potentialApp := range files {
		relevant, err := isPathAnArgoCDApplication(fs, path.Join(absoluteApplicationsDir, potentialApp.Name()))
		if err != nil {
			return nil, fmt.Errorf("checking relevance: %w", err)
		}

		if relevant {
			apps = append(apps, filenameWithoutExtension(potentialApp.Name()))
		}
	}

	return apps, nil
}

func isPathAnArgoCDApplication(fs *afero.Afero, absoluteApplicationPath string) (bool, error) {
	stat, err := fs.Stat(absoluteApplicationPath)
	if err != nil {
		return false, fmt.Errorf("stating file: %w", err)
	}

	if stat.IsDir() {
		return false, nil
	}

	rawFile, err := fs.ReadFile(absoluteApplicationPath)
	if err != nil {
		return false, fmt.Errorf("reading file: %w", err)
	}

	var potentialApp argocd.Application

	err = yaml.Unmarshal(rawFile, &potentialApp)
	if err != nil {
		return false, nil
	}

	return potentialApp.Valid(), nil
}
