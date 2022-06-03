package main

import (
	"context"
	"fmt"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.96.remote-state-versioning/pkg/cfn"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.96.remote-state-versioning/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.96.remote-state-versioning/pkg/lib/logging"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.96.remote-state-versioning/pkg/manifest"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.96.remote-state-versioning/pkg/patch"
	"github.com/spf13/afero"
)

func upgrade(ctx context.Context, log logging.Logger, fs *afero.Afero, flags cmdflags.Flags) error {
	clusterManifest, err := manifest.Cluster(fs)
	if err != nil {
		return fmt.Errorf("acquiring cluster manifest: %w", err)
	}

	stackName := generateStackName(clusterManifest.Metadata.Name)

	log.Debugf("Fetching stack with name %s\n", stackName)

	template, err := cfn.FetchStackTemplate(ctx, stackName)
	if err != nil {
		return fmt.Errorf("fetching template: %w", err)
	}

	log.Debug("Found stack")
	log.Debug("Starting patch operation")

	patchedTemplate, err := patch.AddBucketVersioning(template)
	if err != nil {
		return fmt.Errorf("patching: %w", err)
	}

	log.Debug("Patching successful")
	log.Debug("Updating stack")

	if !flags.DryRun {
		err = cfn.UpdateStackTemplate(ctx, stackName, patchedTemplate)
		if err != nil {
			return fmt.Errorf("updating template: %w", err)
		}
	}

	log.Debug("Update success.")

	return nil
}

func generateStackName(clusterName string) string {
	return fmt.Sprintf("okctl-s3bucket-%s-okctl-%s-meta", clusterName, clusterName)
}
