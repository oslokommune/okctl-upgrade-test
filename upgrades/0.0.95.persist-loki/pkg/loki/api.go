package loki

import (
	"fmt"
	"time"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.95.persist-loki/pkg/lib/context"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.95.persist-loki/pkg/kubectl"
)

// AddPersistence knows how to configure Loki for persistence
func AddPersistence(ctx context.Context, region string, clusterName string, bucketName string) error {
	err := ensureNodeSelector(ctx, "loki", "loki")
	if err != nil {
		return fmt.Errorf("ensuring defined nodeSelector: %w", err)
	}

	// The new config should be active the next calendar preventing issues with index tables spanning two configurations
	from := time.Now().Add(oneDay)

	patch, err := generateLokiPersistencePatch(region, clusterName, bucketName, from)
	if err != nil {
		return fmt.Errorf("generating persistence patch: %w", err)
	}

	original, err := kubectl.GetLokiConfig(ctx.Fs)
	if err != nil {
		return fmt.Errorf("acquiring config: %w", err)
	}

	originalAsJSON, err := asJSON(original)
	if err != nil {
		return fmt.Errorf("converting to JSON: %w", err)
	}

	updated, err := patchConfig(originalAsJSON, patch)
	if err != nil {
		return fmt.Errorf("patching config: %w", err)
	}

	updatedConfigAsYAML, err := asYAML(updated)
	if err != nil {
		return fmt.Errorf("converting to YAML: %w", err)
	}

	if ctx.Flags.DryRun {
		ctx.Logger.Debug("Patching config locally successful. Applying to cluster")

		return nil
	}

	err = kubectl.UpdateLokiConfig(ctx.Fs, updatedConfigAsYAML)
	if err != nil {
		return fmt.Errorf("updating config: %w", err)
	}

	err = kubectl.RestartLoki(ctx.Fs)
	if err != nil {
		return fmt.Errorf("restarting Loki: %w", err)
	}

	return nil
}

func ensureNodeSelector(ctx context.Context, volumeClaimName string, statefulSetName string) error {
	ctx.Logger.Debug("Checking for volume claim")

	exists, err := kubectl.HasVolumeClaim(ctx.Fs, volumeClaimName)
	if err != nil {
		return fmt.Errorf("checking for volume claim: %w", err)
	}

	if !exists {
		ctx.Logger.Debug("Volume claim not found, ignoring node selector")

		return nil
	}

	zone, err := kubectl.GetVolumeClaimAZ(ctx.Fs, volumeClaimName)
	if err != nil {
		return fmt.Errorf("acquiring volume claim AZ: %w", err)
	}

	ctx.Logger.Debugf("Found volume AZ %s, adding node selector to loki", zone)

	if ctx.Flags.DryRun {
		return nil
	}

	err = kubectl.AddNodeSelector(ctx.Fs, statefulSetName, kubectl.AvailabilityZoneLabelKey, zone)
	if err != nil {
		return fmt.Errorf("adding nodeSelector: %w", err)
	}

	return nil
}
