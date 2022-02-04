package argocd

import (
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	merrors "github.com/mishudark/errors"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.87.argocd/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.87.argocd/pkg/lib/logger"
	"github.com/oslokommune/okctl/pkg/api"
	"github.com/oslokommune/okctl/pkg/cfn"
	"github.com/oslokommune/okctl/pkg/client"
	"github.com/oslokommune/okctl/pkg/client/core"
	"github.com/oslokommune/okctl/pkg/config/constant"
	"github.com/oslokommune/okctl/pkg/controller/common/reconciliation"
	"github.com/oslokommune/okctl/pkg/helm/charts/argocd"
	"github.com/oslokommune/okctl/pkg/okctl"
	"strings"
	"time"
)

const (
	argoCDNamespace   = "argocd"
	argoCDIngressName = "argocd-server"

	argoClientSecretName = "argocd/client_secret" //nolint:gosec
	argoSecretKeyName    = "argocd/secret_key"    //nolint:gosec
	argoSecretName       = "argocd-secret"
	argoPrivateKeyName   = "argocd-privatekey"

	deleteReleaseRetrySeconds            = 3
	waitForIngressDeletionTimeoutSeconds = 8 * 60

	appVersionBeforeUpgrade = "v1.6.2"
	appVersionAfterUpgrade  = "v2.1.7"
)

// ArgoCD is a sample okctl component
type ArgoCD struct {
	flags   cmdflags.Flags
	okctl   OkctlTools
	log     logger.Logger
	kubectl Kubectl
}

type OkctlTools struct {
	o         *okctl.Okctl
	clusterID api.ID
	state     *core.StateHandlers
	services  *core.Services
}

var errNothingToDo = errors.New("nothing to do")

// Upgrade upgrades the component
func (a ArgoCD) Upgrade() error {
	err := a.preflight()
	if err != nil {
		if errors.Is(err, errNothingToDo) {
			return nil
		}

		return fmt.Errorf("running preflight checks: %w", err)
	}

	a.log.Infof("Upgrading ArgoCD to version %s\n", appVersionAfterUpgrade)

	err = a.partiallyRemoveArgoCD()
	if err != nil {
		return fmt.Errorf("removing ArgoCD: %w", err)
	}

	err = a.createArgoCD()
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
	if !a.okctl.o.Declaration.Integrations.ArgoCD {
		a.log.Info("ArgoCD is not enabled in cluster declaration, not doing anything")
		return errNothingToDo
	}

	release, err := getHelmRelease(a.okctl.o, argocd.ReleaseName, argocd.Namespace)
	if err != nil {
		if merrors.IsKind(err, merrors.NotExist) {
			return nil
		}

		return fmt.Errorf("getting helm release: %w", err)
	}

	currentVersion := getHelmReleaseAppVersion(release)
	if currentVersion != appVersionBeforeUpgrade {
		a.log.Infof("Current chart version is %s. This upgrade only targets chart version %s. Ignoring upgrade.\n",
			currentVersion, appVersionBeforeUpgrade)
		return errNothingToDo
	}

	return nil
}

func (a ArgoCD) partiallyRemoveArgoCD() error {
	// Delete Helm release so we can reinstall it with correct version
	err := a.deleteHelmReleaseIfExists()
	if err != nil {
		return fmt.Errorf("deleting helm release if exists: %w", err)
	}

	// Delete secrets because their format has changed
	err = a.deleteSecrets()
	if err != nil {
		return fmt.Errorf("deleting secrets: %w", err)
	}

	err = a.waitForIngressToNotExist()
	if err != nil {
		return fmt.Errorf("waiting for ingress to not exist: %w", err)
	}

	return nil
}

func (a ArgoCD) deleteHelmReleaseIfExists() error {
	releaseExists := false
	_, err := getHelmRelease(a.okctl.o, argocd.ReleaseName, argocd.Namespace)
	if err == nil {
		releaseExists = true
	} else if !merrors.IsKind(err, merrors.NotExist) {
		return fmt.Errorf("getting helm release: %w", err)
	}

	if !releaseExists {
		a.log.Info("Helm release doesn't exist, skipping delete")
		return nil
	}

	a.log.Info("Deleting Helm Release")

	if a.flags.DryRun {
		return nil
	}

	err = a.okctl.services.Helm.DeleteHelmRelease(context.Background(), client.DeleteHelmReleaseOpts{
		ID:          a.okctl.clusterID,
		ReleaseName: argocd.ReleaseName,
		Namespace:   argocd.Namespace,
	})
	if err != nil {
		return fmt.Errorf("deleting helm release: %w", err)
	}

	return nil
}

func (a ArgoCD) deleteSecrets() error {
	a.log.Info("Deleting secrets")
	if a.flags.DryRun {
		return nil
	}

	for _, name := range []string{argoSecretName, argoPrivateKeyName} {
		a.log.Debugf("Deleting external secret '%s'\n", name)
		err := a.okctl.services.Manifest.DeleteExternalSecret(context.Background(), client.DeleteExternalSecretOpts{
			ID:   a.okctl.clusterID,
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
		err := a.okctl.services.Parameter.DeleteSecret(context.Background(), client.DeleteSecretOpts{
			ID:   a.okctl.clusterID,
			Name: secret,
		})
		if err != nil {
			return fmt.Errorf("deleting secret: %w", err)
		}
	}

	return nil
}

// The ingress takes a while to delete because the AWS ALB takes some time to remove. If we don't wait for it to be removed,
// the new ingress won't be created.
func (a ArgoCD) waitForIngressToNotExist() error {
	const waitMsg = "Waiting for ArgoCD ingress to disappear"

	a.log.Info(waitMsg)
	if a.flags.DryRun {
		return nil
	}

	secondsPassed := 0

	for {
		hasIngress, err := a.kubectl.hasIngress(argoCDNamespace, argoCDIngressName)
		if err != nil {
			return fmt.Errorf("getting ingress existence: %w", err)
		}

		if hasIngress {
			if secondsPassed > waitForIngressDeletionTimeoutSeconds {
				return errors.New("exceeded timeout for waiting for deletion of ingress")
			}

			a.log.Debug(waitMsg)
			time.Sleep(deleteReleaseRetrySeconds * time.Second)
			secondsPassed += deleteReleaseRetrySeconds
		} else {
			break
		}
	}

	return nil
}

func (a ArgoCD) createArgoCD() error {
	// The following code is mostly copy pasted from argocd_reconciler.go
	hostedZone, err := a.okctl.state.Domain.GetPrimaryHostedZone()
	if err != nil {
		return fmt.Errorf("getting primary hosted zone: %w", err)
	}

	identityPool, err := a.okctl.state.IdentityManager.GetIdentityPool(
		cfn.NewStackNamer().IdentityPool(a.okctl.o.Declaration.Metadata.Name),
	)
	if err != nil {
		return fmt.Errorf("getting identity pool: %w", err)
	}

	a.log.Info("Creating ArgoCD")

	if a.flags.DryRun {
		return nil
	}

	a.log.Debug("CreateGithubRepository")
	repo, err := a.okctl.services.Github.CreateGithubRepository(context.Background(), client.CreateGithubRepositoryOpts{
		ID:           reconciliation.ClusterMetaAsID(a.okctl.o.Declaration.Metadata),
		Host:         constant.DefaultGithubHost,
		Organization: a.okctl.o.Declaration.Github.Organisation,
		Name:         a.okctl.o.Declaration.Github.Repository,
	})
	if err != nil {
		return fmt.Errorf("fetching deploy key: %w", err)
	}

	a.log.Debug("CreateArgoCD")
	_, err = a.okctl.services.ArgoCD.CreateArgoCD(context.Background(), client.CreateArgoCDOpts{
		ID:                 a.okctl.clusterID,
		Domain:             a.okctl.o.Declaration.ClusterRootDomain,
		FQDN:               dns.Fqdn(a.okctl.o.Declaration.ClusterRootDomain),
		HostedZoneID:       hostedZone.HostedZoneID,
		GithubOrganisation: a.okctl.o.Declaration.Github.Organisation,
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

	if a.flags.DryRun {
		return nil
	}

	release, err := getHelmRelease(a.okctl.o, argocd.ReleaseName, argocd.Namespace)
	if err != nil {
		return fmt.Errorf("getting helm release: %w", err)
	}

	currentVersion := getHelmReleaseAppVersion(release)
	if currentVersion != appVersionAfterUpgrade {
		return fmt.Errorf("expected chart version %s, but current version is %s", appVersionAfterUpgrade, currentVersion)
	}

	hasIngress, err := a.kubectl.hasIngress(argoCDNamespace, argoCDIngressName)
	if err != nil {
		return fmt.Errorf("getting ingress existense: %w", err)
	}

	if !hasIngress {
		return fmt.Errorf("expected ArgoCD ingress to be present, but it was not")
	}

	return nil
}

func New(log logger.Logger, flags cmdflags.Flags) (ArgoCD, error) {
	kubectl, err := newKubectl(log)
	if err != nil {
		return ArgoCD{}, fmt.Errorf("creating kubectl: %w", err)
	}

	o, err := initializeOkctl()
	if err != nil {
		return ArgoCD{}, fmt.Errorf("initializing: %w", err)
	}

	state := o.StateHandlers(o.StateNodes())

	services, err := o.ClientServices(state)
	if err != nil {
		return ArgoCD{}, err
	}

	okctlTools := OkctlTools{
		o:         o,
		clusterID: getClusterID(o),
		state:     state,
		services:  services,
	}

	return ArgoCD{
		flags:   flags,
		okctl:   okctlTools,
		log:     log,
		kubectl: kubectl,
	}, nil
}
