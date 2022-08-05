package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func (receiver Client) Apply(manifest io.Reader) error {
	cmd := exec.Command(defaultBinaryName, "apply", "-f", "-") //nolint:gosec

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = manifest
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return nil
}
