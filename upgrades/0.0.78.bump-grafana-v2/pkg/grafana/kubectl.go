package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.78.bump-grafana-v2/pkg/logger"

	"github.com/Masterminds/semver"

	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	grafanaRepository     = "grafana/grafana"
	monitoringNamespace   = "monitoring"
	grafanaDeploymentName = "kube-prometheus-stack-grafana"
	grafanaContainerName  = "grafana"
	timeoutSeconds        = 300
)

var (
	expectedGrafanaVersionPreUpgrade = semver.MustParse("7.3.5")  //nolint:gochecknoglobals
	targetGrafanaVersion             = semver.MustParse("7.5.12") //nolint:gochecknoglobals
)

func acquireKubectlClient(kubeConfigPath string) (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("creating rest config: %w", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("initializing client: %w", err)
	}

	return client, nil
}

func getCurrentGrafanaVersion(clientSet *kubernetes.Clientset) (*semver.Version, error) {
	grafanaContainerIndex, err := getContainerIndexByName(clientSet, grafanaContainerName)
	if err != nil {
		return nil, fmt.Errorf("getting Grafana container index: %w", err)
	}

	result, err := clientSet.AppsV1().Deployments(monitoringNamespace).Get(
		context.Background(),
		grafanaDeploymentName,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	parts := strings.Split(result.Spec.Template.Spec.Containers[grafanaContainerIndex].Image, ":")

	version, err := semver.NewVersion(parts[1])
	if err != nil {
		return nil, fmt.Errorf("parsing version: %w", err)
	}

	return version, nil
}

func patchGrafanaDeployment(log logger.Logger, clientSet *kubernetes.Clientset, dryRun bool) error {
	dryRunOpts := []string{metav1.DryRunAll}
	if !dryRun {
		dryRunOpts = nil
	}

	log.Info("Identifying relevant container")

	grafanaContainerIndex, err := getContainerIndexByName(clientSet, grafanaContainerName)
	if err != nil {
		return fmt.Errorf("acquiring Grafana container index: %w", err)
	}

	log.Debug(fmt.Sprintf("found relevant container at index %d", grafanaContainerIndex))

	log.Info("Generating upgrade patch")

	patch := Patch{
		Op:    jsonPatchOperationReplace,
		Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", grafanaContainerIndex),
		Value: fmt.Sprintf("%s:%s", grafanaRepository, targetGrafanaVersion.String()),
	}

	raw, err := json.Marshal([]Patch{patch})
	if err != nil {
		return fmt.Errorf("marshalling patch: %w", err)
	}

	log.Info("Applying patch")

	_, err = clientSet.AppsV1().Deployments(monitoringNamespace).Patch(
		context.Background(),
		grafanaDeploymentName,
		types.JSONPatchType,
		raw,
		metav1.PatchOptions{DryRun: dryRunOpts},
	)
	if err != nil {
		return fmt.Errorf("patching Grafana deployment: %w", err)
	}

	return nil
}

func hasGrafanaInstalled(clientSet *kubernetes.Clientset) (bool, error) {
	result, err := clientSet.AppsV1().Deployments(monitoringNamespace).List(context.Background(), metav1.ListOptions{
		TimeoutSeconds: int64Ptr(timeoutSeconds),
	})
	if err != nil {
		return false, fmt.Errorf("listing deployments: %w", err)
	}

	for _, deployment := range result.Items {
		if deployment.Name == grafanaDeploymentName {
			return true, nil
		}
	}

	return false, nil
}

func getContainerIndexByName(clientSet *kubernetes.Clientset, name string) (int, error) {
	result, err := clientSet.AppsV1().Deployments(monitoringNamespace).Get(
		context.Background(),
		grafanaDeploymentName,
		metav1.GetOptions{},
	)
	if err != nil {
		return -1, fmt.Errorf("getting deployment data: %w", err)
	}

	for index, container := range result.Spec.Template.Spec.Containers {
		if container.Name == name {
			return index, nil
		}
	}

	return -1, fmt.Errorf("not found")
}
