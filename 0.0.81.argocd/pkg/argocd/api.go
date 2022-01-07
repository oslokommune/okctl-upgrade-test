package argocd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/miekg/dns"
	"github.com/oslokommune/okctl-upgrade/0.0.81.argocd/pkg/logger"
	"github.com/oslokommune/okctl/pkg/api"
	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
	"github.com/oslokommune/okctl/pkg/cfn"
	"github.com/oslokommune/okctl/pkg/client"
	"github.com/oslokommune/okctl/pkg/client/core"
	"github.com/oslokommune/okctl/pkg/config/constant"
	"github.com/oslokommune/okctl/pkg/controller/common/reconciliation"
	"github.com/oslokommune/okctl/pkg/helm/charts/argocd"
	"strings"
)

const (
	argoCDNamespace      = "argocd"
	argoCDDeploymentName = "argocd-application-controller"
	argoCDContainerName  = "application-controller"

	argoClientSecretName = "argocd/client_secret" //nolint:gosec
	argoSecretKeyName    = "argocd/secret_key"    //nolint:gosec
	argoSecretName       = "argocd-secret"
	argoPrivateKeyName   = "argocd-privatekey"
)

var (
	expectedContainerImagePreUpgradeVersion = semver.MustParse("1.7.2")
	targetContainerImageVersion             = semver.MustParse("2.1.7")
)

// ArgoCD is a sample okctl component
type ArgoCD struct {
	log     logger.Logger
	dryRun  bool
	confirm bool
	kubectl Kubectl
}

var errNothingToDo = errors.New("nothing to do")

// Upgrade upgrades the component
func (a ArgoCD) Upgrade() error {
	a.log.Info("Upgrading ArgoCD")

	err := a.preflight()
	if err != nil {
		if errors.Is(err, errNothingToDo) {
			return nil
		}

		return fmt.Errorf("running preflight checks: %w", err)
	}

	o, err := initializeOkctl()
	if err != nil {
		return fmt.Errorf("initializing: %w", err)
	}

	clusterID := getClusterID(o)
	state := o.StateHandlers(o.StateNodes())

	services, err := o.ClientServices(state)
	if err != nil {
		return err
	}

	err = a.partiallyRemoveArgoCD(services, clusterID)
	if err != nil {
		return fmt.Errorf("removing ArgoCD: %w", err)
	}

	err = a.createArgoCD(services, clusterID, state, o.Declaration)
	if err != nil {
		return fmt.Errorf("creating ArgoCD: %w", err)
	}

	err = a.postflight()
	if err != nil {
		return fmt.Errorf("running postflight checks: %w", err)
	}

	a.log.Info("Upgrading ArgoCD done!")

	return nil
}

func (a ArgoCD) preflight() error {
	isInstalled, err := a.kubectl.hasDeployment(argoCDNamespace, argoCDDeploymentName)
	if err != nil {
		return fmt.Errorf("getting argocd: %w", err)
	}

	if !isInstalled {
		a.log.Info("ArgoCD is not installed, not doing anything")
		return errNothingToDo
	}

	currentVersion, err := a.kubectl.getDeploymentImageVersion(argoCDNamespace, argoCDDeploymentName, argoCDContainerName)
	if err != nil {
		return fmt.Errorf("getting current image version: %w", err)
	}

	err = a.validateVersion(expectedContainerImagePreUpgradeVersion, currentVersion)
	if err != nil {
		if currentVersion.GreaterThan(targetContainerImageVersion) || currentVersion.Equal(targetContainerImageVersion) {
			a.log.Infof("Current version is %s, ignoring upgrade\n", currentVersion.String())

			return errNothingToDo
		}

		return fmt.Errorf("unexpected version installed: %w", err)
	}

	return nil
}

func (a ArgoCD) validateVersion(expected *semver.Version, actual *semver.Version) error {
	if !actual.Equal(expected) {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}

	return nil
}

func (a ArgoCD) partiallyRemoveArgoCD(services *core.Services, clusterID api.ID) error {
	if a.dryRun {
		return nil
	}

	a.log.Info("Removing Old ArgoCD")

	a.log.Debug("Deleting Helm Release")
	err := services.Helm.DeleteHelmRelease(context.Background(), client.DeleteHelmReleaseOpts{
		ID:          clusterID,
		ReleaseName: argocd.ReleaseName,
		Namespace:   argocd.Namespace,
	})
	if err != nil {
		return fmt.Errorf("deleting helm release: %w", err)
	}

	for _, name := range []string{argoSecretName, argoPrivateKeyName} {
		a.log.Debugf("Deleting external secret '%s'\n", name)
		err = services.Manifest.DeleteExternalSecret(context.Background(), client.DeleteExternalSecretOpts{
			ID:   clusterID,
			Name: name,
			Secrets: map[string]string{
				name: constant.DefaultArgoCDNamespace,
			},
		})
		if err != nil {
			return fmt.Errorf("deleting external secret: %w", err)
		}
	}

	for _, secret := range []string{argoSecretKeyName, argoClientSecretName} {
		a.log.Debugf("Deleting secret '%s'\n", secret)
		err = services.Parameter.DeleteSecret(context.Background(), client.DeleteSecretOpts{
			ID:   clusterID,
			Name: secret,
		})
		if err != nil {
			return fmt.Errorf("deleting secret: %w", err)
		}
	}

	return nil
}

func (a ArgoCD) createArgoCD(
	services *core.Services,
	clusterID api.ID,
	state *core.StateHandlers,
	declaration *v1alpha1.Cluster,
) error {
	// The following code is mostly copy pasted from argocd_reconciler.go
	hostedZone, err := state.Domain.GetPrimaryHostedZone()
	if err != nil {
		return fmt.Errorf("getting primary hosted zone: %w", err)
	}

	identityPool, err := state.IdentityManager.GetIdentityPool(
		cfn.NewStackNamer().IdentityPool(declaration.Metadata.Name),
	)
	if err != nil {
		return fmt.Errorf("getting identity pool: %w", err)
	}

	if a.dryRun {
		return nil
	}

	a.log.Info("Creating ArgoCD")
	repo, err := services.Github.CreateGithubRepository(context.Background(), client.CreateGithubRepositoryOpts{
		ID:           reconciliation.ClusterMetaAsID(declaration.Metadata),
		Host:         constant.DefaultGithubHost,
		Organization: declaration.Github.Organisation,
		Name:         declaration.Github.Repository,
	})
	if err != nil {
		return fmt.Errorf("fetching deploy key: %w", err)
	}

	_, err = services.ArgoCD.CreateArgoCD(context.Background(), client.CreateArgoCDOpts{
		ID:                 clusterID,
		Domain:             declaration.ClusterRootDomain,
		FQDN:               dns.Fqdn(declaration.ClusterRootDomain),
		HostedZoneID:       hostedZone.HostedZoneID,
		GithubOrganisation: declaration.Github.Organisation,
		UserPoolID:         identityPool.UserPoolID,
		AuthDomain:         identityPool.AuthDomain,
		Repository:         repo,
	})
	if err != nil {
		// nolint: godox
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			a.log.Info(fmt.Errorf("got ArgoCD timeout: %w", err).Error())
		}

		return fmt.Errorf("creating argocd: %w", err)
	}
	return nil
}

func (a ArgoCD) postflight() error {
	a.log.Info("Verifying new ArgoCD version")

	currentVersion, err := a.kubectl.getDeploymentImageVersion(argoCDNamespace, argoCDDeploymentName, argoCDContainerName)
	if err != nil {
		return fmt.Errorf("getting updated image version: %w", err)
	}

	expectedVersion := targetContainerImageVersion
	if a.dryRun {
		expectedVersion = currentVersion
	}

	err = a.validateVersion(expectedVersion, currentVersion)
	if err != nil {
		a.log.Debugf("Expected version %s, but got %s\n", expectedVersion.String(), currentVersion.String())

		return fmt.Errorf("validating new version: %w", err)
	}

	return nil
}

type Opts struct {
	DryRun  bool
	Confirm bool
}

func New(log logger.Logger, opts Opts) (ArgoCD, error) {
	kubectl, err := newKubectl(log)
	if err != nil {
		return ArgoCD{}, fmt.Errorf("creating kubectl: %w", err)
	}

	return ArgoCD{
		log:     log,
		dryRun:  opts.DryRun,
		confirm: opts.Confirm,
		kubectl: kubectl,
	}, nil
}
