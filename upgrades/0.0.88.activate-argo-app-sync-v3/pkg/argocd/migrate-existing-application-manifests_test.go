package argocd

import (
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v3/pkg/logger"
	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
	"github.com/oslokommune/okctl/pkg/jsonpatch"
	"github.com/oslokommune/okctl/pkg/scaffold"
	"github.com/spf13/afero"

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

// Ensures no regression for KM620 / https://trello.com/c/AyCdNq4y where production charts where moved while upgrading
// development environment
func TestEnsureMigratingOnlySingleEnv(t *testing.T) {
	fs := &afero.Afero{Fs: afero.NewMemMapFs()}

	withAppName := "migrate-app"

	withMigratingEnv := v1alpha1.NewCluster()
	withMigratingEnv.Metadata.Name = "development"

	withStaticEnv := v1alpha1.NewCluster()
	withStaticEnv.Metadata.Name = "production"

	err := bootstrapEnvApp(fs, withStaticEnv, withAppName)
	assert.NoError(t, err)

	err = bootstrapEnvApp(fs, withMigratingEnv, withAppName)
	assert.NoError(t, err)

	assertExistence(t, fs, withStaticEnv, withAppName, true)
	assertExistence(t, fs, withMigratingEnv, withAppName, true)

	err = MigrateExistingApplicationManifests(MigrateExistingApplicationManifestsOpts{
		Logger: logger.Logger{
			Out: io.Discard,
			Err: io.Discard,
		},
		DryRun:  false,
		Fs:      fs,
		Cluster: withMigratingEnv,
	})
	assert.NoError(t, err)

	assertExistence(t, fs, withStaticEnv, withAppName, true)
	assertExistence(t, fs, withMigratingEnv, withAppName, false)
}

func bootstrapEnvApp(fs *afero.Afero, env v1alpha1.Cluster, appName string) error {
	appDir := path.Join(env.Github.OutputPath, "applications")
	baseDir := path.Join(appDir, appName, "base")
	overlaysDir := path.Join(appDir, appName, "overlays", env.Metadata.Name)
	app := v1alpha1.Application{Metadata: v1alpha1.ApplicationMeta{Name: appName}}

	err := scaffold.GenerateApplicationBase(scaffold.GenerateApplicationBaseOpts{
		SaveManifest: func(filename string, content []byte) error {
			return fs.WriteFile(path.Join(baseDir, filename), content, 0o600)
		},
		Application: app,
	})
	if err != nil {
		return fmt.Errorf("generating application base: %w", err)
	}

	err = scaffold.GenerateApplicationOverlay(scaffold.GenerateApplicationOverlayOpts{
		SavePatch: func(kind string, patch jsonpatch.Patch) error {
			return nil
		},
		Application:    app,
		Domain:         "mock.domain.ex",
		CertificateARN: "not:an:arn",
	})
	if err != nil {
		return fmt.Errorf("generating application overlays: %w", err)
	}

	err = scaffold.GenerateArgoCDApplicationManifest(scaffold.GenerateArgoCDApplicationManifestOpts{
		Saver: func(content []byte) error {
			return fs.WriteFile(path.Join(overlaysDir, defaultArgoCDApplicationManifestName), content, 0x600)
		},
		Application:                   app,
		IACRepoURL:                    "github.com/org/repo",
		RelativeApplicationOverlayDir: "irrelevantpath/",
	})
	if err != nil {
		return fmt.Errorf("generating ArgoCD application manifest: %w", err)
	}

	return nil
}

func assertExistence(t *testing.T, fs *afero.Afero, env v1alpha1.Cluster, appName string, expectExistence bool) {
	appDir := path.Join(env.Github.OutputPath, "applications")
	targetPath := path.Join(appDir, appName, "overlays", env.Metadata.Name, defaultArgoCDApplicationManifestName)

	exists, err := fs.Exists(targetPath)
	assert.NoError(t, err)

	assert.Equal(t, expectExistence, exists)
}
