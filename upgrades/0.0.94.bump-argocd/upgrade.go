package main

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.bump-argocd/pkg/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.bump-argocd/pkg/lib/cmdflags"
)

var (
	expectedVersion = *semver.MustParse("2.1.7")
	targetVersion   = *semver.MustParse("2.1.15")
)

func upgrade(ctx Context, flags cmdflags.Flags) error {
	argocdServerSelector := kubectl.Selector{
		Namespace:     "argocd",
		Kind:          "deployment",
		Name:          "argocd-server",
		ContainerName: "server",
	}

	exists, err := kubectl.HasResource(argocdServerSelector)
	if err != nil {
		return fmt.Errorf("checking for ArgoCD existence: %w", err)
	}

	if !exists {
		ctx.logger.Debug("ArgoCD not found, nothing to do")

		return nil
	}

	ctx.logger.Debug("Acquiring current ArgoCD version")

	currentVersion, err := kubectl.GetImageVersion(argocdServerSelector)
	if err != nil {
		return fmt.Errorf("acquiring argocd image version: %w", err)
	}

	ctx.logger.Debugf("Found version %s\n", currentVersion.String())

	if !expectedVersion.Equal(&currentVersion) {
		ctx.logger.Debugf("Current version %s does not equal expected version %s, ignoring upgrade",
			currentVersion.String(),
			expectedVersion.String(),
		)

		return nil
	}

	ctx.logger.Debugf("Found expected version, preparing to upgrade\n")

	err = kubectl.UpdateImageVersion(ctx.logger, flags, argocdServerSelector, targetVersion)
	if err != nil {
		return fmt.Errorf("updating version: %w", err)
	}

	return nil
}
