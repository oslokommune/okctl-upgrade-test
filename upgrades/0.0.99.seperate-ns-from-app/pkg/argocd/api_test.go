package argocd

import (
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/sebdah/goldie/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSetupNamespacesSync(t *testing.T) {
	testCases := []struct {
		name        string
		withCluster v1alpha1.Cluster
	}{
		{
			name: "Should setup correct files and directories",
			withCluster: v1alpha1.Cluster{
				Metadata: v1alpha1.ClusterMeta{Name: "mock-cluster"},
				Github: v1alpha1.ClusterGithub{
					Organisation: "mockorg",
					Repository:   "mockrepo",
					OutputPath:   "infrastructure",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := &afero.Afero{Fs: afero.NewMemMapFs()}
			absoluteRepositoryRootDirectory := "/"

			err := EnableNamespacesSync(EnableNamespaceSyncOpts{
				Log:                             &mockDebugLogger{},
				DryRun:                          false,
				Fs:                              fs,
				Kubectl:                         &mockKubectlClient{},
				AbsoluteRepositoryRootDirectory: absoluteRepositoryRootDirectory,
				Cluster:                         tc.withCluster,
			})
			assert.NoError(t, err)

			g := goldie.New(t)

			argocdConfigDir := path.Join(
				tc.withCluster.Github.OutputPath,
				tc.withCluster.Metadata.Name,
				"argocd",
			)

			readmeExists, err := fs.Exists(path.Join(
				absoluteRepositoryRootDirectory,
				argocdConfigDir,
				"namespaces",
				"README.md",
			))
			assert.NoError(t, err)

			assert.True(t, readmeExists)

			argoapp, err := fs.ReadFile(path.Join(absoluteRepositoryRootDirectory, argocdConfigDir, "namespaces.yaml"))
			assert.NoError(t, err)

			g.Assert(t, fmt.Sprintf("argoapp-%s", tc.name), argoapp)
		})
	}
}

type mockDebugLogger struct{}

func (receiver mockDebugLogger) Debug(_ ...interface{}) {}

type mockKubectlClient struct{}

func (m mockKubectlClient) Apply(_ io.Reader) error { return nil }
