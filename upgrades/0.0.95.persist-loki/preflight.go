package main

import (
	"fmt"
	"io"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.95.persist-loki/pkg/apis/okctl.io/v1alpha1"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.95.persist-loki/pkg/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.95.persist-loki/pkg/lib/commonerrors"
	"github.com/spf13/afero"

	"sigs.k8s.io/yaml"
)

func preflight(fs *afero.Afero, clusterManifest v1alpha1.Cluster) error {
	if !clusterManifest.Integrations.Loki {
		return fmt.Errorf("loki integration disabled: %w", commonerrors.ErrNoActionRequired)
	}

	lokiExists, err := kubectl.HasLoki(fs)
	if err != nil {
		return fmt.Errorf("checking loki existence: %w", err)
	}

	if !lokiExists {
		return fmt.Errorf("missing loki installation: %w", commonerrors.ErrNoActionRequired)
	}

	awsStorageExists, err := hasAWSStorageDefined(fs)
	if err != nil {
		return fmt.Errorf("checking for existing AWS configuration: %w", err)
	}

	if awsStorageExists {
		return fmt.Errorf("storage configuration already exists: %w", commonerrors.ErrNoActionRequired)
	}

	return nil
}

func hasAWSStorageDefined(fs *afero.Afero) (bool, error) {
	configStream, err := kubectl.GetLokiConfig(fs)
	if err != nil {
		return false, fmt.Errorf("acquiring config: %w", err)
	}

	rawConfig, err := io.ReadAll(configStream)
	if err != nil {
		return false, fmt.Errorf("buffering: %w", err)
	}

	var cfg config

	err = yaml.Unmarshal(rawConfig, &cfg)
	if err != nil {
		return false, fmt.Errorf("unmarshalling: %w", err)
	}

	_, ok := cfg.StorageConfig["aws"]

	return ok, nil
}

type config struct {
	StorageConfig map[string]interface{} `json:"storage_config"`
}
