// Package v1alpha1 defines common data types and functionality on them
package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

// Parse knows how to convert bytes into a Cluster
func Parse(raw []byte) (Cluster, error) {
	clusterManifest := NewCluster()

	err := yaml.Unmarshal(raw, &clusterManifest)
	if err != nil {
		return Cluster{}, fmt.Errorf("unmarshalling: %w", err)
	}

	return clusterManifest, nil
}
