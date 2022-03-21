package binary

import (
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v3/pkg/logger"
	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
	"github.com/oslokommune/okctl/pkg/binaries"
	"github.com/oslokommune/okctl/pkg/credentials"
	"github.com/spf13/afero"
)

type teardownFn func() error

type client struct {
	logger              logger.Logger
	fs                  *afero.Afero
	binaryProvider      binaries.Provider
	credentialsProvider credentials.Provider
	cluster             v1alpha1.Cluster
}

type NewOpts struct {
	Logger              logger.Logger
	Fs                  *afero.Afero
	BinaryProvider      binaries.Provider
	CredentialsProvider credentials.Provider
	Cluster             v1alpha1.Cluster
}
