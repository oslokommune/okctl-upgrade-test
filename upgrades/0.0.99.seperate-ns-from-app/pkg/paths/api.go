package paths

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
)

const (
	ArgocdConfigDir              = "argocd"
	argocdConfigNamespacesDir    = "namespaces"
	argocdConfigApplicationsDir  = "applications"
	ApplicationsDir              = "applications"
	ApplicationBaseDir           = "base"
	ApplicationOverlaysDir       = "overlays"
	DefaultFolderPermissions     = 0o700
	DefaultReadmeFilename        = "README.md"
	DefaultKustomizationFilename = "kustomization.yaml"
)

func RelativeArgoCDConfigDir(cluster v1alpha1.Cluster) string {
	return path.Join(cluster.Github.OutputPath, cluster.Metadata.Name, ArgocdConfigDir)
}

func RelativeNamespacesDir(cluster v1alpha1.Cluster) string {
	return path.Join(RelativeArgoCDConfigDir(cluster), argocdConfigNamespacesDir)
}

func RelativeArgoCDApplicationsDir(cluster v1alpha1.Cluster) string {
	return path.Join(RelativeArgoCDConfigDir(cluster), argocdConfigApplicationsDir)
}

func RelativeApplicationDir(cluster v1alpha1.Cluster, appName string) string {
	return path.Join(cluster.Github.OutputPath, ApplicationsDir, appName)
}

// GetRepositoryRootDirectory returns the absolute path of the repository root no matter what the current working
// directory of the repository the user is in
func GetRepositoryRootDirectory() (string, error) {
	result, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("getting repository root directory: %w", err)
	}

	pathAsString := string(bytes.Trim(result, "\n"))

	return pathAsString, nil
}
