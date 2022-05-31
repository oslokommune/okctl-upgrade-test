package kubectl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.bump-argocd/pkg/jsonpatch"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.bump-argocd/pkg/kubectl/resources"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.bump-argocd/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.bump-argocd/pkg/lib/logger"
)

type Selector struct {
	Namespace     string
	Kind          string
	Name          string
	ContainerName string
}

func GetImageVersion(selector Selector) (semver.Version, error) {
	if !contains(validKinds, selector.Kind) {
		return semver.Version{}, fmt.Errorf("kind %s is not supported. Valid kinds are %s",
			selector.Kind,
			validKinds,
		)
	}

	result, err := runCommand(
		"--namespace", selector.Namespace,
		"--output", "json",
		"get", selector.Kind, selector.Name,
	)
	if err != nil {
		return semver.Version{}, fmt.Errorf("running command: %w", err)
	}

	raw, err := io.ReadAll(result)
	if err != nil {
		return semver.Version{}, fmt.Errorf("buffering: %w", err)
	}

	var resource resources.Deployment

	err = json.Unmarshal(raw, &resource)
	if err != nil {
		return semver.Version{}, fmt.Errorf("unmarshalling: %w", err)
	}

	var imageURI string

	for _, container := range resource.Spec.Template.Spec.Containers {
		if container.Name == selector.ContainerName {
			imageURI = container.Image
		}
	}

	if imageURI == "" {
		return semver.Version{}, fmt.Errorf("container name %s not found", selector.ContainerName)
	}

	tag := strings.SplitN(imageURI, ":", 2)[1]
	tag = strings.TrimPrefix(tag, "v")

	version, err := semver.NewVersion(tag)
	if err != nil {
		return semver.Version{}, fmt.Errorf("parsing tag: %w", err)
	}

	return *version, nil
}

func UpdateImageVersion(log logger.Logger, flags cmdflags.Flags, selector Selector, newVersion semver.Version) error {
	log.Debug("Acquiring relevant container")

	serverContainerIndex, err := acquireContainerIndex(selector)
	if err != nil {
		return fmt.Errorf("acquiring container index: %w", err)
	}

	log.Debugf("Found relevant container at index %d, generating patch\n", serverContainerIndex)

	patch := jsonpatch.New().Add(jsonpatch.Operation{
		Type:  jsonpatch.OperationTypeReplace,
		Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", serverContainerIndex),
		Value: fmt.Sprintf("quay.io/argoproj/argocd:v%s", newVersion.String()),
	})

	rawPatch, err := patch.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshalling: %w", err)
	}

	log.Debugf("Generated %s\n", rawPatch)

	args := []string{
		"--namespace", selector.Namespace,
		"patch", selector.Kind, selector.Name,
		"--type", "json",
		"--patch", string(rawPatch),
	}

	if flags.DryRun {
		args = append(args, "--dry-run=server")
	}

	stdout, err := runCommand(args...)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}

	rawStdout, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("buffering: %w", err)
	}

	log.Debugf("Kubectl: %s", rawStdout)

	log.Debug("Patching complete, upgrade successful")

	return nil
}

func HasResource(selector Selector) (bool, error) {
	_, err := runCommand(
		"--namespace", selector.Namespace,
		"get", selector.Kind, selector.Name,
	)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("running command: %w", err)
	}

	return true, nil
}

func acquireContainerIndex(selector Selector) (int, error) {
	result, err := runCommand(
		"--namespace", selector.Namespace,
		"--output", "json",
		"get", selector.Kind, selector.Name,
	)
	if err != nil {
		return -1, fmt.Errorf("running command: %w", err)
	}

	raw, err := io.ReadAll(result)
	if err != nil {
		return -1, fmt.Errorf("buffering: %w", err)
	}

	var resource resources.Deployment

	err = json.Unmarshal(raw, &resource)
	if err != nil {
		return -1, fmt.Errorf("unmarshalling: %w", err)
	}

	for index, container := range resource.Spec.Template.Spec.Containers {
		if container.Name == selector.ContainerName {
			return index, nil
		}
	}

	return -1, fmt.Errorf("no container with name %s found", selector.ContainerName)
}

var validKinds = []string{"deployment", "statefulset", "daemonset"}

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if strings.EqualFold(item, needle) {
			return true
		}
	}

	return false
}
