package grafana

import (
	"fmt"

	"github.com/oslokommune/okctl-upgrade/0.0.78.bump-grafana/pkg/logger"
	"k8s.io/client-go/kubernetes"

	"github.com/Masterminds/semver"
)

func validateVersion(expected *semver.Version, actual *semver.Version) error {
	if !actual.Equal(expected) {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}

	return nil
}

func preflight(log logger.Logger, clientSet *kubernetes.Clientset) error {
	hasGrafana, err := hasGrafanaInstalled(clientSet)
	if err != nil {
		return fmt.Errorf("checking Grafana existence: %w", err)
	}

	if !hasGrafana {
		log.Info("Grafana is not installed, ignoring upgrade")

		return ErrNothingToDo
	}

	currentGrafanaVersion, err := getCurrentGrafanaVersion(clientSet)
	if err != nil {
		return fmt.Errorf("getting current Grafana version: %w", err)
	}

	err = validateVersion(expectedGrafanaVersionPreUpgrade, currentGrafanaVersion)
	if err != nil {
		if currentGrafanaVersion.GreaterThan(targetGrafanaVersion) || currentGrafanaVersion.Equal(targetGrafanaVersion) {
			log.Info(fmt.Sprintf("Current version is %s, ignoring upgrade", currentGrafanaVersion.String()))

			return ErrNothingToDo
		}

		return fmt.Errorf("unexpected Grafana version installed: %w", err)
	}

	return nil
}

func postflight(log logger.Logger, clientSet *kubernetes.Clientset, dryRun bool) error {
	log.Info("Verifying new Grafana version")

	newVersion, err := getCurrentGrafanaVersion(clientSet)
	if err != nil {
		return fmt.Errorf("acquiring updated Grafana version: %w", err)
	}

	log.Debug(fmt.Sprintf("Found new Grafana version %s", newVersion.String()))

	expectedVersion := targetGrafanaVersion
	if dryRun {
		expectedVersion = newVersion
	}

	err = validateVersion(expectedVersion, newVersion)
	if err != nil {
		log.Debug("Expected version %s, but got %s", expectedVersion.String(), newVersion.String())

		return fmt.Errorf("validating new version: %w", err)
	}

	return nil
}

func int64Ptr(i int64) *int64 {
	return &i
}
