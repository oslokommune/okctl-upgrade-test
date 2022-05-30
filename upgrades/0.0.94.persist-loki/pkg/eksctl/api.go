// Package eksctl defines a simplified API for dealing with eks related operations
package eksctl

import (
	"fmt"
	"os"
	"strings"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/context"
)

// CreateServiceAccount knows how to create a service account -> IAM role mapping, with a set of policies
func CreateServiceAccount(ctx context.Context, clusterName string, name string, policies []string) error {
	eksctlPath, err := acquireEksctlPath(ctx.Fs, os.UserHomeDir)
	if err != nil {
		return fmt.Errorf("acquiring eksctl path: %w", err)
	}

	args := []string{
		"create", "iamserviceaccount",
		"--name", name,
		"--namespace", defaultMonitoringNamespace,
		"--cluster", clusterName,
		"--role-name", fmt.Sprintf("okctl-%s-loki", clusterName),
		"--attach-policy-arn", strings.Join(policies, ","),
		"--approve",
		"--override-existing-serviceaccounts",
	}

	if ctx.Flags.DryRun {
		ctx.Logger.Debugf("Running eksctl with args: %v\n", args)

		return nil
	}

	err = runEksctlCommand(eksctlPath, args...)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}

	return nil
}
