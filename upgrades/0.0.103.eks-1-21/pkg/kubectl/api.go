package kubectl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func (kubectlClient) GetVersion() (Versions, error) {
	cmd := exec.Command(defaultBinaryName, "version", "-o", "yaml") //nolint:gosec

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		return Versions{}, fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return ParseVersion(stdout.Bytes())
}

func New() Client {
	return kubectlClient{}
}
