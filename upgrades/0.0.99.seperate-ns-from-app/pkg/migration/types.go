package migration

import (
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
)

type debugLogger interface {
	Debug(args ...interface{})
}

type MigrateExistingApplicationNamespacesToClusterOpts struct {
	Log                    debugLogger
	DryRun                 bool
	Fs                     *afero.Afero
	Cluster                v1alpha1.Cluster
	AbsoluteRepositoryRoot string
}

type kustomization struct {
	Resources []string `json:"resources"`
}
