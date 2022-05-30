package kubectl

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/afero"
)

func getVolumeID(fs *afero.Afero, claimName string) (string, error) {
	output, err := runCommand(fs,
		"--namespace", defaultMonitoringNamespace,
		"--output", "json",
		"get", persistentVolumeClaimResourceKind, claimName,
	)
	if err != nil {
		return "", fmt.Errorf("retrieving claim: %w", err)
	}

	rawResponse, err := io.ReadAll(output)
	if err != nil {
		return "", fmt.Errorf("buffering: %w", err)
	}

	var response pvc

	err = json.Unmarshal(rawResponse, &response)
	if err != nil {
		return "", fmt.Errorf("unmarshalling: %w", err)
	}

	return response.Spec.VolumeName, nil
}

func getVolumeZone(fs *afero.Afero, volumeName string) (string, error) {
	output, err := runCommand(fs,
		"--namespace", defaultMonitoringNamespace,
		"--output", "json",
		"get", persistentVolumeResourceKind, volumeName,
	)
	if err != nil {
		return "", fmt.Errorf("retrieving volume: %w", err)
	}

	rawOutput, err := io.ReadAll(output)
	if err != nil {
		return "", fmt.Errorf("buffering: %w", err)
	}

	var volume pv

	err = json.Unmarshal(rawOutput, &volume)
	if err != nil {
		return "", fmt.Errorf("unmarshalling: %w", err)
	}

	return volume.Metadata.Labels[AvailabilityZoneLabelKey], nil
}
