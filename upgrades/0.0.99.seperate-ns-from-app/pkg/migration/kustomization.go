package migration

import (
	"bytes"
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

func deleteNamespaceEntryFromKustomizationResources(fs *afero.Afero, kustomizationPath string) error {
	rawKustomizationFile, err := fs.ReadFile(kustomizationPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var kustomizationFile kustomization

	err = yaml.Unmarshal(rawKustomizationFile, &kustomizationFile)
	if err != nil {
		return fmt.Errorf("unmarshalling: %w", err)
	}

	namespaceResourceIndex := -1

	for index, item := range kustomizationFile.Resources {
		if item == "namespace.yaml" {
			namespaceResourceIndex = index

			break
		}
	}

	if namespaceResourceIndex == -1 {
		return nil
	}

	kustomizationFile.Resources = append(
		kustomizationFile.Resources[0:namespaceResourceIndex],
		kustomizationFile.Resources[namespaceResourceIndex+1:]...,
	)

	rawUpdatedKustomization, err := yaml.Marshal(kustomizationFile)
	if err != nil {
		return fmt.Errorf("marshalling: %w", err)
	}

	err = fs.WriteReader(kustomizationPath, bytes.NewReader(rawUpdatedKustomization))
	if err != nil {
		return fmt.Errorf("writing updated file: %w", err)
	}

	return nil
}
