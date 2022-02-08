package argocd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// getAllArgoCDApplicationManifests walks rootDir recursively and returns all files named 'argocd-application.yaml'
func getAllArgoCDApplicationManifests(filesystem *afero.Afero, rootDir string) ([]string, error) {
	accumulator := pathAccumulator{Paths: make([]string, 0)}

	err := filesystem.Walk(rootDir, generateAppManifestWalker(&accumulator))
	if err != nil {
		return nil, fmt.Errorf("gathering ArgoCD application manifests: %w", err)
	}

	return accumulator.Paths, nil
}

func generateAppManifestWalker(accumulator *pathAccumulator) filepath.WalkFunc {
	return func(currentPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking app manifests: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() != defaultArgoCDApplicationManifestName {
			return nil
		}

		accumulator.Push(currentPath)

		return nil
	}
}

// getApplicationNameFromPath knows how to extract the application name from the path of an ArgoCD application manifest
func getApplicationNameFromPath(applicationsRootDirectory string, targetPath string) string {
	cleanedPath := strings.Replace(targetPath, applicationsRootDirectory, "", 1)
	cleanedPath = strings.TrimPrefix(cleanedPath, "/")

	parts := strings.Split(cleanedPath, "/")

	return parts[0]
}

type pathAccumulator struct {
	Paths []string
}

func (receiver *pathAccumulator) Push(newPath string) {
	receiver.Paths = append(receiver.Paths, newPath)
}
