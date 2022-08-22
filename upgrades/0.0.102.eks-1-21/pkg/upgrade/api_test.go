package upgrade_test

import (
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/kubectl/version"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/lib/cmdflags"
	"testing"

	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/lib/logging"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.102.eks-1-21/pkg/upgrade"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	testCases := []struct {
		name          string
		version       string
		errorContains string
	}{
		{
			name:          "Should return error",
			version:       "20+",
			errorContains: "current EKS version is 1.20, but must be at least 1.21",
		},
		{
			version:       "21+",
			name:          "Should return error nothing to do",
			errorContains: "nothing to do",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kubectlClient := mockKubectlWithVersion(tc.version)
			flags := cmdflags.Flags{}
			logger := logging.New(logging.Debug)

			err := upgrade.Start(logger, flags, kubectlClient)

			assert.Error(t, err)
			assert.Containsf(t, err.Error(), tc.errorContains, "unexpected error message")
		})
	}
}

type KubectlClientMock struct {
	serverMinorVersion string
}

func (m KubectlClientMock) GetVersion() (version.Versions, error) {
	return version.Versions{
		ClientVersion: version.Version{},
		ServerVersion: version.Version{
			Minor: m.serverMinorVersion,
		},
	}, nil
}

func mockKubectlWithVersion(version string) kubectl.Client {
	return KubectlClientMock{
		serverMinorVersion: version,
	}
}
