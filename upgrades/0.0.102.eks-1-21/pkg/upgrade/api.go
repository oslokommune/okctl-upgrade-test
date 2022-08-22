package upgrade

import (
	"fmt"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/lib/cmdflags"

	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/lib/commonerrors"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/lib/logging"
)

const minimumEKSMinorVersion = 21

func Start(logger logging.Logger, flags cmdflags.Flags, kubectl kubectl.Client) error {
	version, err := kubectl.GetVersion()
	if err != nil {
		return fmt.Errorf("running kubectl version: %w", err)
	}

	currentEKSMinorVersion, err := version.ServerVersion.MinorAsInt()
	if err != nil {
		return fmt.Errorf("getting EKS minor version as an integer: %w", err)
	}

	if currentEKSMinorVersion >= minimumEKSMinorVersion {
		logger.Debugf("Not doing anything, as this upgrade targets EKS minor version %d and below,"+
			" and current EKS minor version is already %d.\n",
			minimumEKSMinorVersion-1, currentEKSMinorVersion)
		return commonerrors.ErrNothingToDo
	}

	logger.Info("")
	logger.Info("❗ IMPORTANT")
	logger.Info("")
	logger.Infof("Current EKS version is: 1.%d\n", currentEKSMinorVersion)
	logger.Info("You must upgrade EKS to version 1.21 by following this guide: " +
		"https://github.com/oslokommune/okctl-upgrade/tree/main/gists/bump-eks")
	logger.Info("")
	logger.Info("For more information, see https://www.okctl.io/eks-cluster-upgrades")
	logger.Info("")

	if flags.DryRun {
		return nil
	} else {
		return fmt.Errorf("current EKS version is 1.%d, but must be at least 1.%d",
			currentEKSMinorVersion, minimumEKSMinorVersion)
	}
}
