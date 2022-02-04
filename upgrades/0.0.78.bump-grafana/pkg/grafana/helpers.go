package grafana

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.78.bump-grafana/pkg/commonerrors"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.78.bump-grafana/pkg/logger"
	"k8s.io/client-go/kubernetes"

	"github.com/Masterminds/semver"
)

func validateVersion(expected *semver.Version, actual *semver.Version) error {
	if !actual.Equal(expected) {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}

	return nil
}

func (c Upgrader) preflight(clientSet *kubernetes.Clientset) error {
	hasGrafana, err := hasGrafanaInstalled(clientSet)
	if err != nil {
		return fmt.Errorf("checking Grafana existence: %w", err)
	}

	if !hasGrafana {
		c.logger.Info("Grafana is not installed, ignoring upgrade")

		return ErrNothingToDo
	}

	currentGrafanaVersion, err := getCurrentGrafanaVersion(clientSet)
	if err != nil {
		return fmt.Errorf("getting current Grafana version: %w", err)
	}

	err = validateVersion(expectedGrafanaVersionPreUpgrade, currentGrafanaVersion)
	if err != nil {
		if currentGrafanaVersion.GreaterThan(targetGrafanaVersion) || currentGrafanaVersion.Equal(targetGrafanaVersion) {
			c.logger.Info(fmt.Sprintf("Current version is %s, ignoring upgrade", currentGrafanaVersion.String()))

			return ErrNothingToDo
		}

		return fmt.Errorf("unexpected Grafana version installed: %w", err)
	}

	c.showWarningMessage()

	if !c.dryRun && !c.confirm {
		yes, err := c.askUser("Do you want to continue?")
		if err != nil {
			return fmt.Errorf("prompting user: %w", err)
		}

		if !yes {
			return commonerrors.ErrUserAborted
		}
	}

	return nil
}

func (c Upgrader) showWarningMessage() {
	if c.dryRun {
		c.logger.Infof(`


***** WARNING *****
During this upgrade, Grafana needs to restart, and all user data stored in Grafana including dashboards, WILL be lost ` +
			`(unless you have added configuration to prevent this)!

For more details and possible mitigations, see: https://www.okctl.io/new-upgrade-for-grafana-available

`)
	} else {
		c.logger.Infof(`



***** WARNING *****
During this upgrade, Grafana needs to restart, and all user data stored in Grafana including dashboards, WILL be lost ` +
			`(unless you have added configuration to prevent this)! Logs and metrics will not be affected, as these are ` +
			`stored in Loki and Prometheus, respectively.

If you have made no adjustments to Grafana after the initial setup of Okctl, you can safely continue with this upgrade.

For more details and possible mitigations, see: https://www.okctl.io/new-upgrade-for-grafana-available

`)
	}
}

func (c Upgrader) askUser(question string) (bool, error) {
	answer := false
	prompt := &survey.Confirm{
		Message: question,
	}

	err := survey.AskOne(prompt, &answer)
	if err != nil {
		return false, err
	}

	return answer, nil
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
