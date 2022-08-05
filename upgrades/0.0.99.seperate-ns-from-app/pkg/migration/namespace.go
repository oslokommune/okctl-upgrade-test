package migration

import (
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

type namespaceMetadata struct {
	Name string
}

type namespaceManifest struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   namespaceMetadata `json:"metadata"`
}

func getNamespaceName(fs *afero.Afero, targetPath string) (string, error) {
	raw, err := fs.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("reading: %w", err)
	}

	var ns namespaceManifest

	err = yaml.Unmarshal(raw, &ns)
	if err != nil {
		return "", fmt.Errorf("unmarshalling: %w", err)
	}

	return ns.Metadata.Name, nil
}
