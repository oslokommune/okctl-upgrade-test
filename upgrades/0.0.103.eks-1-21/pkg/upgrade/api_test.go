package upgrade_test

import (
	"testing"

	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.103.eks-1-21/pkg/kubectl"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.103.eks-1-21/pkg/lib/logging"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.103.eks-1-21/pkg/upgrade"
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
			logger := logging.New(logging.Debug)

			err := upgrade.Start(logger, kubectlClient)

			assert.Error(t, err)
			assert.Containsf(t, err.Error(), tc.errorContains, "unexpected error message")
		})
	}
}

type KubectlClientMock struct {
	serverMinorVersion string
}

func (m KubectlClientMock) GetVersion() (kubectl.Versions, error) {
	return kubectl.Versions{
		ClientVersion: kubectl.Version{},
		ServerVersion: kubectl.Version{
			Minor: m.serverMinorVersion,
		},
	}, nil
}

func mockKubectlWithVersion(version string) kubectl.Client {
	return KubectlClientMock{
		serverMinorVersion: version,
	}
}
