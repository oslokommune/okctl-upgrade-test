package migration

import (
	"io"
	"path"
	"strings"
	"testing"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/paths"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type testFile struct {
	path    string
	content io.Reader
}

type testCluster struct {
	name    string
	apps    []string
	current bool
}

func TestMigrateExistingApplicationNamespacesToCluster(t *testing.T) {
	testCases := []struct {
		name                   string
		withClusters           []testCluster
		withFiles              []testFile
		expectExistingFiles    []string
		expectNonExistingFiles []string
	}{
		{
			name: "Should move the namespace manifest and clean up in a single cluster setup",
			withClusters: []testCluster{
				{name: "cluster", apps: []string{"hello"}, current: true},
			},
			withFiles: []testFile{
				{path: "/infrastructure/applications/hello/base/namespace.yaml", content: namespace("namespace-one")},
			},
			expectExistingFiles: []string{
				"/infrastructure/cluster/argocd/namespaces/namespace-one.yaml",
			},
			expectNonExistingFiles: []string{
				"/infrastructure/applications/hello/base/namespace.yaml",
			},
		},
		{
			name: "Should copy the namespace manifest and leave the original in a multi cluster setup with unmigrated other dependent cluster",
			withClusters: []testCluster{
				{name: "cluster", apps: []string{"hello"}, current: true},
				{name: "cluster-two", apps: []string{"hello"}},
			},
			withFiles: []testFile{
				{path: "/infrastructure/applications/hello/base/namespace.yaml", content: namespace("namespace-one")},
			},
			expectExistingFiles: []string{
				"/infrastructure/applications/hello/base/namespace.yaml",
				"/infrastructure/cluster/argocd/namespaces/namespace-one.yaml",
			},
			expectNonExistingFiles: []string{
				"/infrastructure/cluster-two/argocd/namespaces/namespace-one.yaml",
			},
		},
		{
			name: "Should copy namespace resource and clean up original upon migration of the final cluster",
			withClusters: []testCluster{
				{name: "cluster", apps: []string{"hello"}, current: true},
				{name: "cluster-two", apps: []string{"hello"}},
			},
			withFiles: []testFile{
				{path: "/infrastructure/applications/hello/base/namespace.yaml", content: namespace("namespace-one")},
				{path: "/infrastructure/cluster-two/argocd/namespaces/namespace-one.yaml", content: namespace("namespace-one")},
			},
			expectExistingFiles: []string{
				"/infrastructure/cluster/argocd/namespaces/namespace-one.yaml",
				"/infrastructure/cluster-two/argocd/namespaces/namespace-one.yaml",
			},
			expectNonExistingFiles: []string{
				"/infrastructure/applications/hello/base/namespace.yaml",
			},
		},
		{
			name: "Should copy namespace resource and clean up in app with multiple clusters, but only one dependent cluster",
			withClusters: []testCluster{
				{name: "cluster", apps: []string{"hello"}, current: true},
				{name: "cluster-two", apps: []string{}},
			},
			withFiles: []testFile{
				{path: "/infrastructure/applications/hello/base/namespace.yaml", content: namespace("namespace-one")},
			},
			expectExistingFiles: []string{
				"/infrastructure/cluster/argocd/namespaces/namespace-one.yaml",
			},
			expectNonExistingFiles: []string{
				"/infrastructure/applications/hello/base/namespace.yaml",
				"/infrastructure/cluster-two/argocd/namespaces/namespace-one.yaml",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := &afero.Afero{Fs: afero.NewMemMapFs()}
			var currentCluster string

			for _, cluster := range tc.withClusters {
				if cluster.current {
					currentCluster = cluster.name
				}

				createCluster(t, fs, cluster.name)

				for _, app := range cluster.apps {
					createApp(t, fs, app)
					addAppToCluster(t, fs, app, cluster.name)
				}
			}

			for _, item := range tc.withFiles {
				err := fs.MkdirAll(path.Dir(item.path), paths.DefaultFolderPermissions)
				assert.NoError(t, err)

				content := item.content

				if content == nil {
					content = strings.NewReader("")
				}

				err = fs.WriteReader(item.path, content)
				assert.NoError(t, err)
			}

			err := MigrateExistingApplicationNamespacesToCluster(MigrateExistingApplicationNamespacesToClusterOpts{
				Log:                    mockDebugLogger{},
				DryRun:                 false,
				Fs:                     fs,
				Cluster:                mockCluster(currentCluster),
				AbsoluteRepositoryRoot: "/",
			})
			assert.NoError(t, err)

			for _, item := range tc.expectExistingFiles {
				exists, err := fs.Exists(item)
				assert.NoError(t, err)

				assert.True(t, exists)
			}

			for _, item := range tc.expectNonExistingFiles {
				exists, err := fs.Exists(item)
				assert.NoError(t, err)

				assert.False(t, exists)
			}
		})
	}
}
