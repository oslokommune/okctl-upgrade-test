package version_test

import (
	_ "embed"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/kubectl/version"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/kubectlVersion.yaml
var kubectlVersionYaml string

func TestParseVersion(t *testing.T) {
	testCases := []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "Should parse server version",
			yaml: kubectlVersionYaml,
			want: "19+",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			yamlAsBytes := []byte(tc.yaml)

			got, err := version.ParseVersions(yamlAsBytes)
			require.NoError(t, err)

			assert.Equal(t, tc.want, got.ServerVersion.Minor)
		})
	}
}

func TestVersion_MinorAsInt(t *testing.T) {
	testCases := []struct {
		name string
		yaml string
		want int
	}{
		{
			name: "Should parse server version",
			yaml: kubectlVersionYaml,
			want: 19,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			yamlAsBytes := []byte(tc.yaml)

			got, err := version.ParseVersions(yamlAsBytes)
			require.NoError(t, err)

			minor, err := got.ServerVersion.MinorAsInt()
			require.NoError(t, err)

			assert.Equal(t, tc.want, minor)
		})
	}
}
