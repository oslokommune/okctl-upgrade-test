package argocd

import (
	"path"

	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
)

// getArgoCDClusterConfigDir returns the cluster specific configuration folder for ArgoCD.
// I.e.: /infrastructure/<cluster name>/argocd
func getArgoCDClusterConfigDir(cluster v1alpha1.Cluster) string {
	return path.Join(
		cluster.Github.OutputPath,
		cluster.Metadata.Name,
		defaultArgoCDClusterConfigDirName,
	)
}
