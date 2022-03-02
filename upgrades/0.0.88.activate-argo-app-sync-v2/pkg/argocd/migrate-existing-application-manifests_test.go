package argocd

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppNameFromPath(t *testing.T) {
	baseDir := path.Join("infrastructure", "applications")

	testCases := []struct {
		name string

		withPath   string
		expectName string
	}{
		{
			name: "Should work",

			withPath:   path.Join(baseDir, "testApp", "overlays", "fancy-cluster", "argocd-application.yaml"),
			expectName: "testApp",
		},
		{
			name: "Should work",

			withPath:   path.Join(baseDir, "frontend", "overlays", "somecluster", "argocd-application.yaml"),
			expectName: "frontend",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			result := getApplicationNameFromPath(baseDir, tc.withPath)

			assert.Equal(t, tc.expectName, result)
		})
	}
}
