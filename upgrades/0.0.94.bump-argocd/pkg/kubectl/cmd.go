package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func runCommand(args ...string) (io.Reader, error) {
	cmd := exec.Command(defaultBinaryName, args...) //nolint:gosec

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%s: %w", stderr.String(), err)

		if isErrNotFound(err) {
			return nil, errNotFound
		}

		return nil, fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return &stdout, nil
}
