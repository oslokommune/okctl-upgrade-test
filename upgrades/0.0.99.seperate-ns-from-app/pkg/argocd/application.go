package argocd

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"text/template"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.99.seperate-ns-from-app/pkg/lib/manifest/apis/okctl.io/v1alpha1"
)

const (
	argoCDApplicationAPIVersion = "argoproj.io/v1alpha1"
	argoCDApplicationKind       = "Application"
)

type Application struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

func (receiver Application) Valid() bool {
	if receiver.ApiVersion != argoCDApplicationAPIVersion {
		return false
	}

	if receiver.Kind != argoCDApplicationKind {
		return false
	}

	return true
}

func scaffoldApplication(cluster v1alpha1.Cluster, name string, targetDir string) (io.Reader, error) {
	t, err := template.New("argo-app").Parse(argoCDApplicationTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}

	buf := bytes.Buffer{}

	err = t.Execute(&buf, struct {
		Name          string
		TargetDir     string
		RepositoryURI string
	}{
		Name:          name,
		TargetDir:     targetDir,
		RepositoryURI: cluster.Github.URL(),
	})
	if err != nil {
		return nil, fmt.Errorf("executing: %w", err)
	}

	return &buf, nil
}

//go:embed templates/argocd-application.yaml
var argoCDApplicationTemplate string
