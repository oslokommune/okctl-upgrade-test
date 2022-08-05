package migration

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestDeleteNamespaceEntryFromKustomizationResources(t *testing.T) {
	testCases := []struct {
		name          string
		withEntries   []string
		expectEntries []string
	}{
		{
			name:          "Should work for a single entry",
			withEntries:   []string{"namespace.yaml"},
			expectEntries: []string{},
		},
		{
			name:          "Should work for multiple entries and namespace.yaml at the start",
			withEntries:   []string{"namespace.yaml", "service.yaml", "deployment.yaml"},
			expectEntries: []string{"service.yaml", "deployment.yaml"},
		},
		{
			name:          "Should work for multiple entries and namespace.yaml not at the start",
			withEntries:   []string{"service.yaml", "namespace.yaml", "deployment.yaml"},
			expectEntries: []string{"service.yaml", "deployment.yaml"},
		},
		{
			name:          "Should work for multiple entries and namespace.yaml at the end",
			withEntries:   []string{"service.yaml", "deployment.yaml", "namespace.yaml"},
			expectEntries: []string{"service.yaml", "deployment.yaml"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := &afero.Afero{Fs: afero.NewMemMapFs()}

			originalKustomization := kustomization{Resources: tc.withEntries}
			kustomizationPath := "/kustomization.yaml"

			rawOriginalKustomizationFile, err := yaml.Marshal(originalKustomization)
			assert.NoError(t, err)

			err = fs.WriteReader(kustomizationPath, bytes.NewReader(rawOriginalKustomizationFile))
			assert.NoError(t, err)

			err = deleteNamespaceEntryFromKustomizationResources(fs, kustomizationPath)
			assert.NoError(t, err)

			rawUpdatedKustomizationFile, err := fs.ReadFile(kustomizationPath)
			assert.NoError(t, err)

			var updatedKustomization kustomization

			err = yaml.Unmarshal(rawUpdatedKustomizationFile, &updatedKustomization)
			assert.NoError(t, err)

			assert.Equal(t, len(tc.expectEntries), len(updatedKustomization.Resources))

			for _, item := range tc.expectEntries {
				assert.True(t, contains(updatedKustomization.Resources, item))
			}

			for _, item := range updatedKustomization.Resources {
				assert.True(t, contains(tc.expectEntries, item))
			}
		})
	}
}
