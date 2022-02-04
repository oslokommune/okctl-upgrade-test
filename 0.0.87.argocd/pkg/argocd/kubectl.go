package argocd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"k8s.io/api/networking/v1"
	"os"
	"strings"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/oslokommune/okctl-upgrade/0.0.87.argocd/pkg/lib/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	timeoutSeconds = 300
)

type Kubectl struct {
	logger    logger.Logger
	clientSet *kubernetes.Clientset
}

func (k Kubectl) hasDeployment(namespace, deployment string) (bool, error) {
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

func (k Kubectl) getDeploymentImageVersion(namespace, deployment, container string) (*semver.Version, error) {
	containerIndex, err := k.getContainerIndexByName(namespace, deployment, container)
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

func (k Kubectl) getContainerIndexByName(namespace, deployment, container string) (int, error) {
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

	return -1, fmt.Errorf("not found")
}

func (k Kubectl) getIngress(namespace, ingress string) (*v1.Ingress, error) {
	i, err := k.clientSet.NetworkingV1().Ingresses(namespace).Get(context.Background(), ingress, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting ingress (namespace: %s, ingress: %s): %w", namespace, ingress, err)
	}

	return i, nil
}

func (k Kubectl) hasIngress(namespace, ingress string) (bool, error) {
	_, err := k.getIngress(namespace, ingress)
	if err != nil {
		return false, nil
	}

	return true, nil
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

func newKubectl(logger logger.Logger) (Kubectl, error) {
	clientSet, err := acquireKubectlClient()
	if err != nil {
		return Kubectl{}, fmt.Errorf("aqcuiring kubectl client: %w", err)
	}

	return Kubectl{
		logger:    logger,
		clientSet: clientSet,
	}, nil
}
