package upgrade

import (
	"context"
	"errors"
	"testing"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/commonerrors"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/logging"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	testCases := []struct {
		name      string
		withFs    *afero.Afero
		expectErr error
	}{
		{
			name: "Should do nothing upon missing applications folder",
			withFs: func() *afero.Afero {
				fs := &afero.Afero{Fs: afero.NewMemMapFs()}

				_ = fs.MkdirAll("/infrastructure/mock-cluster-prod/argocd", 0o700)

				return fs
			}(),
			expectErr: commonerrors.ErrNothingToDo,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := v1alpha1.Cluster{
				Metadata: v1alpha1.ClusterMeta{Name: "mock-cluster"},
				Github:   v1alpha1.ClusterGithub{OutputPath: "infrastructure"},
			}

			flags := cmdflags.Flags{
				Debug:   false,
				DryRun:  false,
				Confirm: true,
			}

			err := Start(context.Background(), logging.Logger{}, tc.withFs, flags, cluster)
			assert.True(t, errors.Is(err, tc.expectErr))
		})
	}
}
