package memlimit

import (
	"errors"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/cmdflags"
	kubectlPkg "github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/logger"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Increaser increases argocd-repo-server's memory limit
type Increaser struct {
	flags   cmdflags.Flags
	log     logger.Logger
	kubectl kubectlPkg.Kubectl
}

const (
	argoCDNamespace                = "argocd"
	argoCDRepoServerDeploymentName = "argocd-repo-server"
	argoCDRepoServerContainerName  = "repo-server"

	jsonPatchOperationReplace = "replace"
)

// errNothingToDo signals that there is nothing to do
var errNothingToDo = errors.New("nothing to do")

// Patch is a JSON patch: https://datatracker.ietf.org/doc/html/rfc6902
type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

var (
	expectedExistingMemoryLimit = resource.MustParse("256Mi") //nolint:gochecknoglobals
	newMemoryLimit              = resource.MustParse("512Mi") //nolint:gochecknoglobals
)
