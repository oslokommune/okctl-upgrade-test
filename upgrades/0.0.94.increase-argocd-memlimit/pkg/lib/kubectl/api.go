// Package kubectl contains functionality for interacting with Kubernetes
package kubectl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94-increase-argocd-memlimit/pkg/lib/logger"
	appsV1 "k8s.io/api/apps/v1"
	containerV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HasDeployment returns true if the deployment exists
func (k Kubectl) HasDeployment(namespace, deployment string) (bool, error) {
	result, err := k.clientSet.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{
		TimeoutSeconds: int64Ptr(timeoutSeconds),
	})
	if err != nil {
		return false, fmt.Errorf("listing deployments: %w", err)
	}

	for _, d := range result.Items {
		if d.Name == deployment {
			return true, nil
		}
	}

	return false, nil
}

// GetDeployment returns a deployment resource
func (k Kubectl) GetDeployment(namespace, name string) (*appsV1.Deployment, error) {
	deployment, err := k.clientSet.AppsV1().Deployments(namespace).Get(
		context.Background(),
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	return deployment, nil
}

// GetContainer returns a container resource
func (k Kubectl) GetContainer(namespace, deploymentName, containerName string) (*containerV1.Container, error) {
	deployment, err := k.GetDeployment(namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("getting deployment '%s': %w", deployment, err)
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			return &container, nil
		}
	}

	return nil, ErrNotFound
}

// GetDeploymentImageVersion returns the version of a deployment image
func (k Kubectl) GetDeploymentImageVersion(namespace, deployment, container string) (*semver.Version, error) {
	containerIndex, err := k.GetContainerIndexByName(namespace, deployment, container)
	if err != nil {
		return nil, fmt.Errorf("getting container index: %w", err)
	}

	result, err := k.clientSet.AppsV1().Deployments(namespace).Get(
		context.Background(),
		deployment,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	parts := strings.Split(result.Spec.Template.Spec.Containers[containerIndex].Image, ":")

	version, err := semver.NewVersion(parts[1])
	if err != nil {
		return nil, fmt.Errorf("parsing version: %w", err)
	}

	return version, nil
}

// GetContainerIndexByName returns the array index of a container in a deployment
func (k Kubectl) GetContainerIndexByName(namespace, deployment, container string) (int, error) {
	result, err := k.clientSet.AppsV1().Deployments(namespace).Get(
		context.Background(),
		deployment,
		metav1.GetOptions{},
	)
	if err != nil {
		return -1, fmt.Errorf("getting deployment data: %w", err)
	}

	for index, c := range result.Spec.Template.Spec.Containers {
		if c.Name == container {
			return index, nil
		}
	}

	return -1, ErrNotFound
}

// GetIngress returns an ingress
func (k Kubectl) GetIngress(namespace, ingress string) (*v1.Ingress, error) {
	i, err := k.clientSet.NetworkingV1().Ingresses(namespace).Get(context.Background(), ingress, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting ingress (namespace: %s, ingress: %s): %w", namespace, ingress, err)
	}

	return i, nil
}

// HasIngress returns true if the specified ingress exists
func (k Kubectl) HasIngress(namespace, ingress string) (bool, error) {
	_, err := k.GetIngress(namespace, ingress)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// PatchDeployment patches a deployment
func (k Kubectl) PatchDeployment(namespace string, deploymentName string, patchJSON []byte, dryRun bool) error {
	var dryRunOpts []string
	if dryRun {
		dryRunOpts = []string{metav1.DryRunAll}
	}

	_, err := k.clientSet.AppsV1().Deployments(namespace).Patch(
		context.Background(),
		deploymentName,
		types.JSONPatchType,
		patchJSON,
		metav1.PatchOptions{DryRun: dryRunOpts},
	)
	if err != nil {
		return fmt.Errorf(": %w", err)
	}

	return nil
}

func int64Ptr(i int64) *int64 {
	return &i
}

func acquireKubectlClient() (*kubernetes.Clientset, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		return nil, errors.New("missing required KUBECONFIG environment variable")
	}

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

// New returns an instance of Kubectl
func New(log logger.Logger) (Kubectl, error) {
	clientSet, err := acquireKubectlClient()
	if err != nil {
		return Kubectl{}, fmt.Errorf("aqcuiring kubectl client: %w", err)
	}

	return Kubectl{
		log:       log,
		clientSet: clientSet,
	}, nil
}
