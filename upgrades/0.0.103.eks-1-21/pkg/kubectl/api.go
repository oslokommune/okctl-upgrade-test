package kubectl

import (
	"bytes"
	"fmt"
	"github.com/oslokommune/okctl-upgrade/upgrades/okctl-upgrade/upgrades/0.0.103.eks-1-21/pkg/kubectl/version"
	"os"
	"os/exec"
)

func (kubectlClient) GetVersion() (version.Versions, error) {
	cmd := exec.Command(defaultBinaryName, "version", "-o", "yaml") //nolint:gosec

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		return version.Versions{}, fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return version.ParseVersions(stdout.Bytes())
}

func New() Client {
	return kubectlClient{}
}
