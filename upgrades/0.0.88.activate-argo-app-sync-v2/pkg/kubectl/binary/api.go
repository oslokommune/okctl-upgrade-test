package binary

import (
	"fmt"
	"io"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v2/pkg/kubectl"
)

// Apply runs kubectl apply on a manifest
func (c client) Apply(manifest io.Reader) error {
	targetPath, teardowner, err := c.cacheReaderOnFs(manifest)
	if err != nil {
		return fmt.Errorf("caching manifest on file system: %w", err)
	}

	defer func() {
		_ = teardowner()
	}()

	err = c.applyFile(targetPath)
	if err != nil {
		return fmt.Errorf("applying manifest: %w", err)
	}

	return nil
}

// New returns an initialized kubectl binary client
func New(opts NewOpts) kubectl.Client {
	return &client{
		logger:              opts.Logger,
		fs:                  opts.Fs,
		binaryProvider:      opts.BinaryProvider,
		credentialsProvider: opts.CredentialsProvider,
		cluster:             opts.Cluster,
	}
}
