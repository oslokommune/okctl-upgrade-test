package eksctl

import (
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestAcquireEksctlPath(t *testing.T) {
	testCases := []struct {
		name         string
		withVersions []string
		expectedPath string
	}{
		{
			name:         "Should work with a single version",
			withVersions: []string{"0.0.90"},
			expectedPath: "/home/gopher/.okctl/binaries/eksctl/0.0.90/linux/amd64/eksctl",
		},
		{
			name:         "Should work with multiple versions",
			withVersions: []string{"0.0.90", "1.20.3", "1.3.4"},
			expectedPath: "/home/gopher/.okctl/binaries/eksctl/1.20.3/linux/amd64/eksctl",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := &afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := path.Join("/", "home", "gopher")
			rootDir := path.Join(homeDir, ".okctl", "binaries", "eksctl")
			postfix := path.Join("linux", "amd64")
			homeDirFn := func() (string, error) {
				return homeDir, nil
			}

			for _, version := range tc.withVersions {
				err := fs.MkdirAll(path.Join(
					rootDir,
					version,
					postfix,
				), 0o600)
				assert.NoError(t, err)
			}

			actualPath, err := acquireEksctlPath(fs, homeDirFn)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedPath, actualPath)
		})
	}
}
