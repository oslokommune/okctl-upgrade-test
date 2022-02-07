package rotatesshkey

import (
	"errors"
	"fmt"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/lib/logger"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/lib/okctlimport"
	"github.com/oslokommune/okctl/pkg/api"
	"github.com/oslokommune/okctl/pkg/client/core"
	"github.com/oslokommune/okctl/pkg/okctl"
)

// SshKeyRotater is a sample okctl component
type SshKeyRotater struct {
	flags cmdflags.Flags
	log   logger.Logger
	okctl OkctlTools
}

var errNothingToDo = errors.New("nothing to do")

// Upgrade upgrades the component
func (r SshKeyRotater) Upgrade() error {
	r.log.Info("Rotating ArgoCD SSH keys using format ed25519")

	if r.flags.DryRun {
		r.log.Info("Simulating some stuff")
	} else {
		r.log.Info("Doing some stuff")
	}

	r.log.Info("Rotating ArgoCD SSH keys done!")

	return nil
}

type OkctlTools struct {
	o         *okctl.Okctl
	clusterID api.ID
	state     *core.StateHandlers
	services  *core.Services
}

func New(logger logger.Logger, flags cmdflags.Flags) (SshKeyRotater, error) {
	o, err := okctlimport.InitializeOkctl()
	if err != nil {
		return SshKeyRotater{}, fmt.Errorf("initializing: %w", err)
	}

	state := o.StateHandlers(o.StateNodes())

	services, err := o.ClientServices(state)
	if err != nil {
		return SshKeyRotater{}, err
	}

	okctlTools := OkctlTools{
		o:         o,
		clusterID: okctlimport.GetClusterID(o),
		state:     state,
		services:  services,
	}

	return SshKeyRotater{
		log:   logger,
		flags: flags,
		okctl: okctlTools,
	}, nil
}
