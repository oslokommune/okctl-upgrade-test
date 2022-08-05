package migration

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/paths"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func mockCluster(name string) v1alpha1.Cluster {
	return v1alpha1.Cluster{
		Metadata: v1alpha1.ClusterMeta{Name: name},
		Github:   v1alpha1.ClusterGithub{OutputPath: "infrastructure"},
	}
}

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}

const appTemplate = `apiVersion: argoproj.io/v1alpha1
kind: Application
`

func TestScanForRelevantApps(t *testing.T) {
	testCases := []struct {
		name        string
		withCluster v1alpha1.Cluster
		withFs      *afero.Afero
		expectApps  []string
	}{
		{
			name:        "Should return correct apps",
			withCluster: mockCluster("mock-cluster"),
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				_ = fs.MkdirAll("/infrastructure/mock-cluster/argocd/applications", paths.DefaultFolderPermissions)
				_ = fs.WriteReader("/infrastructure/mock-cluster/argocd/applications/mock-app-one.yaml", strings.NewReader(appTemplate))

				return fs
			}(),
			expectApps: []string{"mock-app-one"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			apps, err := getApplicationsInCluster(tc.withFs, tc.withCluster, "/")
			assert.NoError(t, err)

			assert.Equal(t, len(tc.expectApps), len(apps))

			for _, item := range apps {
				assert.True(t, contains(tc.expectApps, item))
			}

			for _, item := range tc.expectApps {
				assert.True(t, contains(apps, item))
			}
		})
	}
}

const namespaceTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{.Name}}`

func namespace(name string) io.Reader {
	t := template.Must(template.New("namespace").Parse(namespaceTemplate))

	buf := bytes.Buffer{}

	_ = t.Execute(&buf, struct {
		Name string
	}{Name: name})

	return &buf
}

func addOldAppNamespace(t *testing.T, fs *afero.Afero, appName string, namespaceName string) {
	appBaseDir := path.Join("/infrastructure/applications", appName, "base")

	err := fs.MkdirAll(appBaseDir, paths.DefaultFolderPermissions)
	assert.NoError(t, err)

	err = fs.WriteReader(path.Join(appBaseDir, "namespace.yaml"), namespace(namespaceName))
	assert.NoError(t, err)
}

func TestMigrateApplication(t *testing.T) {
	testCases := []struct {
		name             string
		withFs           *afero.Afero
		withCluster      v1alpha1.Cluster
		withAppName      string
		expectNamespaces []string
	}{
		{
			name: "Should successfully migrate app with old namespace",
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				addOldAppNamespace(t, fs, "mock-app-one", "mock-namespace")

				return fs
			}(),
			withCluster:      mockCluster("mock-cluster"),
			withAppName:      "mock-app-one",
			expectNamespaces: []string{"mock-namespace"},
		},
		{
			name: "Should do nothing when theres no namespace manifest in base to be found",
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				addOldAppNamespace(t, fs, "mock-app-one", "mock-namespace")
				_ = fs.Remove("/infrastructure/applications/mock-app-one/base/namespace.yaml")

				return fs
			}(),
			withCluster:      mockCluster("mock-cluster"),
			withAppName:      "mock-app-one",
			expectNamespaces: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := migrateApplication(&mockDebugLogger{}, false, tc.withFs, tc.withCluster, "/", tc.withAppName)
			assert.NoError(t, err)

			for _, ns := range tc.expectNamespaces {
				exists, err := tc.withFs.Exists(path.Join(
					"/",
					tc.withCluster.Github.OutputPath,
					tc.withCluster.Metadata.Name,
					"argocd",
					"namespaces",
					fmt.Sprintf("%s.yaml", ns),
				))
				assert.NoError(t, err)

				assert.True(t, exists)
			}
		})
	}
}

const kustomizeTemplate = `
resources:
- deployment.yaml
- service.yaml
- namespace.yaml
- ingress.yaml
`

// One app "exists" when the /infrastructure/applications/<app name> directory exists
func createApp(t *testing.T, fs *afero.Afero, appName string) {
	absOutputDir := path.Join("/", "infrastructure")
	absAppDir := path.Join(absOutputDir, "applications", appName)
	absBaseDir := path.Join(absAppDir, "base")

	err := fs.MkdirAll(absBaseDir, paths.DefaultFolderPermissions)
	assert.NoError(t, err)

	err = fs.WriteReader(path.Join(absBaseDir, "kustomization.yaml"), strings.NewReader(kustomizeTemplate))
	assert.NoError(t, err)
}

// One cluster "exists" when the /infrastructure/<cluster name> directory exists
func createCluster(t *testing.T, fs *afero.Afero, clusterName string) {
	absOutputDir := path.Join("/", "infrastructure")
	absArgoCDApplicationsConfigDir := path.Join(absOutputDir, clusterName, "argocd", "applications")

	err := fs.MkdirAll(absArgoCDApplicationsConfigDir, paths.DefaultFolderPermissions)
	assert.NoError(t, err)
}

const argoCDApplicationTemplate = `apiVersion: argoproj.io/v1alpha1
kind: Application
`

// An app is added to a cluster when the /infrastructure/<cluster name>/argocd/applications/<app name>.yaml file exists
func addAppToCluster(t *testing.T, fs *afero.Afero, appName string, clusterName string) {
	absOutputDir := path.Join("/", "infrastructure")
	absArgoCDApplicationsConfigDir := path.Join(absOutputDir, clusterName, "argocd", "applications")

	err := fs.MkdirAll(absArgoCDApplicationsConfigDir, paths.DefaultFolderPermissions)
	assert.NoError(t, err)

	err = fs.WriteReader(
		path.Join(absArgoCDApplicationsConfigDir, fmt.Sprintf("%s.yaml", appName)),
		strings.NewReader(argoCDApplicationTemplate),
	)
	assert.NoError(t, err)

	absAppClusterOverlaysDir := path.Join(absOutputDir, "applications", appName, "overlays", clusterName)
	err = fs.MkdirAll(absAppClusterOverlaysDir, paths.DefaultFolderPermissions)
	assert.NoError(t, err)
}

func addNewAppNamespace(t *testing.T, fs *afero.Afero, clusterName string, namespaceName string) {
	absOutputDir := path.Join("/", "infrastructure")
	absArgoCDNamespacesConfigDir := path.Join(absOutputDir, clusterName, "argocd", "namespaces")

	err := fs.MkdirAll(absArgoCDNamespacesConfigDir, paths.DefaultFolderPermissions)
	assert.NoError(t, err)

	err = fs.WriteReader(
		path.Join(absArgoCDNamespacesConfigDir, fmt.Sprintf("%s.yaml", namespaceName)),
		namespace(namespaceName),
	)
	assert.NoError(t, err)
}

func TestRemoveRedundantNamespacesFromBase(t *testing.T) {
	testCases := []struct {
		name                     string
		withFs                   *afero.Afero
		expectedNonExistantPaths []string
		expectedExistantPaths    []string
		withCurrentCluster       v1alpha1.Cluster
	}{
		{
			name:               "Should remove base ns with one app and one upgraded cluster",
			withCurrentCluster: mockCluster("mock-prod"),
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				clusterOne := "mock-prod"
				createCluster(t, fs, clusterOne)

				appOne := "mock-app-one"
				createApp(t, fs, appOne)

				appOneNamespace := "apps"
				addOldAppNamespace(t, fs, appOne, appOneNamespace)
				addNewAppNamespace(t, fs, clusterOne, appOneNamespace)

				addAppToCluster(t, fs, appOne, clusterOne)

				return fs
			}(),
			expectedNonExistantPaths: []string{"/infrastructure/applications/mock-app-one/base/namespace.yaml"},
			expectedExistantPaths:    []string{"/infrastructure/mock-prod/argocd/namespaces/apps.yaml"},
		},
		{
			name:               "Should leave ns in base with one upgraded cluster and one not upgraded cluster",
			withCurrentCluster: mockCluster("mock-prod"),
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				clusterOne := "mock-prod"
				createCluster(t, fs, clusterOne)

				clusterTwo := "mock-test"
				createCluster(t, fs, clusterTwo)

				appOne := "mock-app-one"
				createApp(t, fs, appOne)

				appOneNamespace := "apps"
				addOldAppNamespace(t, fs, appOne, appOneNamespace)
				addNewAppNamespace(t, fs, clusterOne, appOneNamespace)

				addAppToCluster(t, fs, appOne, clusterOne)
				addAppToCluster(t, fs, appOne, clusterTwo)

				return fs
			}(),
			expectedNonExistantPaths: []string{},
			expectedExistantPaths: []string{
				"/infrastructure/mock-prod/argocd/namespaces/apps.yaml",
				"/infrastructure/applications/mock-app-one/base/namespace.yaml",
			},
		},
		{
			name: "Should remove namespace from app even with another irrelevant app has a namespace of the same name",
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				clusterOne := "mock-cluster-one"
				createCluster(t, fs, clusterOne)

				clusterTwo := "mock-cluster-two"
				createCluster(t, fs, clusterTwo)

				appOne := "mock-app-one"
				createApp(t, fs, appOne)

				appTwo := "mock-app-two"
				createApp(t, fs, "mock-app-two")

				appOneNamespace := "mock-namespace-one"
				addOldAppNamespace(t, fs, appOne, appOneNamespace)
				addNewAppNamespace(t, fs, clusterOne, appOneNamespace)

				appTwoNamespace := appOneNamespace
				addOldAppNamespace(t, fs, appTwo, appTwoNamespace)

				addAppToCluster(t, fs, appOne, clusterOne)
				addAppToCluster(t, fs, appTwo, clusterTwo)

				return fs
			}(),
			expectedNonExistantPaths: []string{
				"/infrastructure/applications/mock-app-one/base/namespace.yaml",
				"/infrastructure/mock-cluster-two/argocd/namespaces/apps.yaml",
			},
			expectedExistantPaths: []string{
				"/infrastructure/applications/mock-app-two/base/namespace.yaml",
			},
			withCurrentCluster: mockCluster("mock-cluster-one"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := removeRedundantNamespacesFromBase(&mockDebugLogger{}, false, tc.withFs, tc.withCurrentCluster, "/")
			assert.NoError(t, err)

			for _, currentPath := range tc.expectedExistantPaths {
				exists, err := tc.withFs.Exists(currentPath)
				assert.NoError(t, err)

				assert.True(t, exists)
			}

			for _, currentPath := range tc.expectedNonExistantPaths {
				exists, err := tc.withFs.Exists(currentPath)
				assert.NoError(t, err)

				assert.False(t, exists)
			}
		})
	}
}

func TestGetApplicationsInCluster(t *testing.T) {
	testCases := []struct {
		name              string
		withFs            *afero.Afero
		withClusterName   string
		expectedAppsFound []string
	}{
		{
			name:            "Should work",
			withClusterName: "mock-cluster",
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				createCluster(t, fs, "mock-cluster")
				createApp(t, fs, "mock-app-one")

				addAppToCluster(t, fs, "mock-app-one", "mock-cluster")

				return fs
			}(),
			expectedAppsFound: []string{"mock-app-one"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			apps, err := getApplicationsInCluster(
				tc.withFs,
				mockCluster(tc.withClusterName),
				"/",
			)
			assert.NoError(t, err)

			assert.Equal(t, len(tc.expectedAppsFound), len(apps))

			for _, app := range apps {
				assert.True(t, contains(tc.expectedAppsFound, app))
			}

			for _, app := range tc.expectedAppsFound {
				assert.True(t, contains(apps, app))
			}
		})
	}
}

func TestEnsureKustomizeIsUpdatedWhenRemovingNamespace(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "Should not find a namespace entry in kustomization.yaml file",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := &afero.Afero{Fs: afero.NewMemMapFs()}

			const clusterOne = "mock-cluster"
			createCluster(t, fs, clusterOne)

			const appOne = "mock-app-one"
			createApp(t, fs, appOne)
			addAppToCluster(t, fs, appOne, clusterOne)

			const namespaceOne = "mock-namespace-one"
			addOldAppNamespace(t, fs, appOne, namespaceOne)
			addNewAppNamespace(t, fs, clusterOne, namespaceOne)

			err := MigrateExistingApplicationNamespacesToCluster(MigrateExistingApplicationNamespacesToClusterOpts{
				Log:                    &mockDebugLogger{},
				DryRun:                 false,
				Fs:                     fs,
				Cluster:                mockCluster(clusterOne),
				AbsoluteRepositoryRoot: "/",
			})
			assert.NoError(t, err)

			rawKustomization, err := fs.ReadFile(path.Join(
				"/",
				"infrastructure",
				"applications",
				appOne,
				"base",
				"kustomization.yaml",
			))
			assert.NoError(t, err)

			var baseKustomization kustomization

			err = yaml.Unmarshal(rawKustomization, &baseKustomization)
			assert.NoError(t, err)

			assert.False(t, contains(baseKustomization.Resources, "namespace.yaml"))
		})
	}
}

type mockDebugLogger struct{}

func (receiver mockDebugLogger) Debug(_ ...interface{}) {}
