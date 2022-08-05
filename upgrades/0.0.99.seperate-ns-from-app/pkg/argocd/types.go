package argocd

import (
	"io"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
)

type debugLogger interface {
	Debug(args ...interface{})
}

type applier interface {
	Apply(io.Reader) error
}

type EnableNamespaceSyncOpts struct {
	Log                             debugLogger
	DryRun                          bool
	Fs                              *afero.Afero
	Kubectl                         applier
	AbsoluteRepositoryRootDirectory string
	Cluster                         v1alpha1.Cluster
}
