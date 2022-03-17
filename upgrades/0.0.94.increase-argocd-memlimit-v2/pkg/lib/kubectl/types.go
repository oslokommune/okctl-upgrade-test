package kubectl

import (
	"errors"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.increase-argocd-memlimit-v2/pkg/lib/logger"
	"k8s.io/client-go/kubernetes"
)

const (
	timeoutSeconds = 300
)

// Kubectl provides functionality for Kubernetes
type Kubectl struct {
	log       logger.Logger
	clientSet *kubernetes.Clientset
}

// ErrNotFound signals that some entity was not found
var ErrNotFound = errors.New("not found")
