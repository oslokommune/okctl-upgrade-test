package kubectl

import "github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.103.eks-1-21/pkg/kubectl/version"

type Client interface {
	GetVersion() (version.Versions, error)
}

type kubectlClient struct{}
