package argocd

import (
	"bytes"
	"fmt"
	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
	"io"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func readClusterDeclaration(path string) (*v1alpha1.Cluster, error) {
	inputReader, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	var (
		buffer  bytes.Buffer
		cluster v1alpha1.Cluster
	)

	cluster = v1alpha1.NewCluster()

	_, err = io.Copy(&buffer, inputReader)
	if err != nil {
		return nil, fmt.Errorf("copying reader data: %w", err)
	}

	err = yaml.Unmarshal(buffer.Bytes(), &cluster)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling buffer: %w", err)
	}

	return &cluster, nil
}
