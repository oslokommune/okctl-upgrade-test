package eksctl

import (
	"bytes"
	"errors"
	"fmt"
	fileSystem "io/fs"
	"os/exec"
	"path"
	"runtime"

	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
)

func acquireEksctlPath(fs *afero.Afero, homeDirFn func() (string, error)) (string, error) {
	homeDir, err := homeDirFn()
	if err != nil {
		return "", fmt.Errorf("acquiring home directory: %w", err)
	}

	eksctlDir := path.Join(homeDir, defaultOkctlConfigDirName, defaultOkctlBinariesDirName, defaultEksctlName)

	exists, err := fs.DirExists(eksctlDir)
	if err != nil {
		return "", fmt.Errorf("checking eksctl directory existence: %w", err)
	}

	if !exists {
		return "", errors.New("missing eksctl directory")
	}

	versions, err := gatherVersions(fs, eksctlDir)
	if err != nil {
		return "", fmt.Errorf("gathering versions: %w", err)
	}

	var latest semver.Version
	for version := range versions {
		if version.GreaterThan(&latest) {
			latest = version
		}
	}

	return path.Join(eksctlDir, latest.String(), runtime.GOOS, runtime.GOARCH, defaultEksctlName), nil
}

func runEksctlCommand(binaryPath string, args ...string) error {
	cmd := exec.Command(binaryPath, args...) //nolint:gosec

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return nil
}

func gatherVersions(fs *afero.Afero, baseDir string) (map[semver.Version]interface{}, error) {
	versions := make(map[semver.Version]interface{})

	err := fs.Walk(baseDir, func(currentPath string, info fileSystem.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isDirectory, err := fs.IsDir(currentPath)
		if err != nil {
			return fmt.Errorf("checking path type: %w", err)
		}

		if !isDirectory {
			return nil
		}

		version, err := semver.NewVersion(path.Base(currentPath))
		if err == nil {
			versions[*version] = true
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return versions, nil
}
