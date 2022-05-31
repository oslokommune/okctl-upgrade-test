package main

import (
	"fmt"
	"os"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.95.persist-loki/pkg/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
)

const clusterManifestPathEnvKey = "OKCTL_CLUSTER_DECLARATION"

func getManifest(fs *afero.Afero) (v1alpha1.Cluster, error) {
	clusterManifestPath := os.Getenv(clusterManifestPathEnvKey)

	rawManifest, err := fs.ReadFile(clusterManifestPath)
	if err != nil {
		return v1alpha1.Cluster{}, fmt.Errorf("reading manifest: %w", err)
	}

	clusterManifest, err := v1alpha1.Parse(rawManifest)
	if err != nil {
		return v1alpha1.Cluster{}, fmt.Errorf("parsing manifest: %w", err)
	}

	return clusterManifest, nil
}
