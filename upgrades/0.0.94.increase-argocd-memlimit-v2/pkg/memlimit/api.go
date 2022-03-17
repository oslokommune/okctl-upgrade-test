// Package memlimit increases the memory limit of argocd-repo-server-deployment
package memlimit

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/cmdflags"
	kubectlPkg "github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/logger"
)

// Upgrade upgrades the component
func (c Increaser) Upgrade() error {
	c.log.Info("Increasing memory limit on ArgoCD's deployment argocd-repo-server from 256Mi to 512Mi")

	err := c.preflight()
	if err != nil {
		if errors.Is(err, errNothingToDo) {
			return nil
		}

		return fmt.Errorf("running preflight checks: %w", err)
	}

	err = c.patchArgoCD()
	if err != nil {
		return fmt.Errorf("patching ArgoCD: %w", err)
	}

	c.log.Info("Upgrading ArgoCD done!")

	return nil
}

func (c Increaser) preflight() error {
	hasDeployment, err := c.kubectl.HasDeployment(argoCDNamespace, argoCDRepoServerDeploymentName)
	if err != nil {
		return fmt.Errorf("checking deployment '%s' existence: %w", argoCDRepoServerDeploymentName, err)
	}

	if !hasDeployment {
		c.log.Info("ArgoCD is not installed, ignoring upgrade")

		return errNothingToDo
	}

	container, err := c.kubectl.GetContainer(argoCDNamespace, argoCDRepoServerDeploymentName, argoCDRepoServerContainerName)
	if err != nil {
		return fmt.Errorf("getting current Grafana version: %w", err)
	}

	existingMemoryLimit := container.Resources.Limits.Memory()
	if !existingMemoryLimit.Equal(expectedExistingMemoryLimit) {
		c.log.Infof("Deployment argocd-repo-server existing CPU limit is %s, but expected is %s, not doing anything.\n",
			existingMemoryLimit.String(), expectedExistingMemoryLimit.String())

		return errNothingToDo
	}

	return nil
}

func (c Increaser) patchArgoCD() error {
	c.log.Debug("Identifying relevant container")

	containerIndex, err := c.kubectl.GetContainerIndexByName(
		argoCDNamespace, argoCDRepoServerDeploymentName, argoCDRepoServerContainerName)
	if err != nil {
		return fmt.Errorf("acquiring container index: %w", err)
	}

	c.log.Debugf("Found relevant container at index %d\n", containerIndex)

	c.log.Debug("Generating upgrade patch")

	patch := Patch{
		Op:    jsonPatchOperationReplace,
		Path:  fmt.Sprintf("/spec/template/spec/containers/%d/resources/limits/memory", containerIndex),
		Value: newMemoryLimit.String(),
	}

	patchJSON, err := json.Marshal([]Patch{patch})
	if err != nil {
		return fmt.Errorf("marshalling patch: %w", err)
	}

	c.log.Info("Applying patch")

	err = c.kubectl.PatchDeployment(argoCDNamespace, argoCDRepoServerDeploymentName, patchJSON, c.flags.DryRun)
	if err != nil {
		return fmt.Errorf("patching deployment: %w", err)
	}

	return nil
}

// New returns a new Increaser
func New(log logger.Logger, flags cmdflags.Flags) (Increaser, error) {
	kubectl, err := kubectlPkg.New(log)
	if err != nil {
		return Increaser{}, fmt.Errorf("creating kubectl: %w", err)
	}

	return Increaser{
		log:     log,
		flags:   flags,
		kubectl: kubectl,
	}, nil
}
