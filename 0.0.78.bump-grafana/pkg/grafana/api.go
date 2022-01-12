package grafana

import (
	"errors"
	"fmt"
	"os"

	"github.com/oslokommune/okctl-upgrade/0.0.78.bump-grafana/pkg/logger"
)

// Upgrader is a sample okctl component
type Upgrader struct {
	logger  logger.Logger
	dryRun  bool
	confirm bool
}

// Upgrade upgrades the component
func (c Upgrader) Upgrade() error {
	c.logger.Info("Upgrading Grafana")

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		return errors.New("missing required KUBECONFIG environment variable")
	}

	kubectlClient, err := acquireKubectlClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("acquiring kubectl client: %w", err)
	}

	if c.dryRun {
		c.logger.Info("Simulating upgrade")
	} else {
		c.logger.Info("Patching Grafana")
	}

	err = c.preflight(kubectlClient)
	if err != nil {
		if errors.Is(err, ErrNothingToDo) {
			return nil
		}

		return fmt.Errorf("running preflight checks: %w", err)
	}

	c.logger.Debug(fmt.Sprintf("Passed preflight test. Upgrading Grafana to %s", targetGrafanaVersion.String()))

	err = patchGrafanaDeployment(c.logger, kubectlClient, c.dryRun)
	if err != nil {
		return fmt.Errorf("patching grafana deployment: %w", err)
	}

	err = postflight(c.logger, kubectlClient, c.dryRun)
	if err != nil {
		return fmt.Errorf("running postflight checks: %w", err)
	}

	c.logger.Info("Upgrading Grafana done!")

	return nil
}

type Opts struct {
	DryRun  bool
	Confirm bool
}

func New(logger logger.Logger, opts Opts) Upgrader {
	return Upgrader{
		logger:  logger,
		dryRun:  opts.DryRun,
		confirm: opts.Confirm,
	}
}
