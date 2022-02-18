package rotatesshkey

import (
	"context"
	"errors"
	"fmt"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/cmdflags"
	upgradeGithub "github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/github"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/logger"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/okctlinit"
	"github.com/oslokommune/okctl/pkg/api"
	"github.com/oslokommune/okctl/pkg/api/core"
	awsProvider "github.com/oslokommune/okctl/pkg/api/core/cloudprovider/aws"
	"github.com/oslokommune/okctl/pkg/apis/okctl.io/v1alpha1"
	"github.com/oslokommune/okctl/pkg/client"
	clientCore "github.com/oslokommune/okctl/pkg/client/core"
	"github.com/oslokommune/okctl/pkg/config/constant"
	"github.com/oslokommune/okctl/pkg/github"
)

// SSHKeyRotater is a sample okctl component
type SSHKeyRotater struct {
	flags                 cmdflags.Flags
	log                   logger.Logger
	clusterID             api.ID
	declaration           *v1alpha1.Cluster
	githubState           client.GithubState
	parameterService      api.ParameterService
	githubDeployKeyGetter *upgradeGithub.Github
	githubClient          *github.Github
	githubService         client.GithubService
}

// Upgrade upgrades the component
func (r SSHKeyRotater) Upgrade() error {
	r.log.Info("Rotating ArgoCD SSH keys")

	err := r.removeDeployKeysIfExists()
	if err != nil {
		return fmt.Errorf("removing deploy key: %w", err)
	}

	err = r.createDeployKey()
	if err != nil {
		return fmt.Errorf("creating deploy key: %w", err)
	}

	r.log.Info("Rotating ArgoCD SSH keys done!")

	return nil
}

func (r SSHKeyRotater) removeDeployKeysIfExists() error {
	// Because the github deploy key ID in our state can be wrong, we cannot use okctl's implemented functionality to remove the
	// github deploy key, because it requires the correct ID. We have to get the correct ID by getting the deploy key by expected
	// name.
	existingKeysToCleanup, err := r.getGithubDeployKeys()
	if err != nil && !errors.Is(err, upgradeGithub.ErrNotFound) {
		return fmt.Errorf("getting github deploy key identifier: %w", err)
	}

	r.log.Infof("Found %d old deploy keys to remove from GitHub\n", len(existingKeysToCleanup))

	for _, key := range existingKeysToCleanup {
		r.log.Infof("Deleting old deploy key %s (id: %d)\n", key.GetTitle(), key.GetID())

		if r.flags.DryRun {
			continue
		}

		err = r.githubClient.DeleteDeployKey(
			r.declaration.Github.Organisation, r.declaration.Github.Repository, key.GetID())
		if err != nil && !errors.Is(err, upgradeGithub.ErrNotFound) {
			return fmt.Errorf("deleting deploy key in GitHub: %w", err)
		}
	}

	return nil
}

func (r SSHKeyRotater) getGithubDeployKeys() ([]*upgradeGithub.Key, error) {
	deployKeyTitle := fmt.Sprintf("okctl-iac-%s", r.declaration.Metadata.Name)

	keys, err := r.githubDeployKeyGetter.GetDeployKeys(
		r.declaration.Github.Organisation, r.declaration.Github.Repository, deployKeyTitle)
	if err != nil {
		return nil, fmt.Errorf("getting deploy key '%s': %w", deployKeyTitle, err)
	}

	for _, key := range keys {
		if key.GetID() == 0 {
			return nil, fmt.Errorf("received deploy key '%s' without ID", key.GetTitle())
		}
	}

	return keys, nil
}

func (r SSHKeyRotater) createDeployKey() error {
	r.log.Info("Creating deploy key")

	if r.flags.DryRun {
		return nil
	}

	deployKey, err := r.githubService.CreateRepositoryDeployKey(client.CreateGithubDeployKeyOpts{
		ID:           r.clusterID,
		Organisation: r.declaration.Github.Organisation,
		Repository:   r.declaration.Github.Repository,
		Title:        fmt.Sprintf("okctl-iac-%s", r.clusterID.ClusterName),
	})
	if err != nil {
		return fmt.Errorf("creating repository deploy key: %w", err)
	}

	r.log.Debugf("New public key: %s\n", deployKey.PublicKey)

	fullName := fmt.Sprintf("%s/%s", r.declaration.Github.Organisation, r.declaration.Github.Repository)

	repo := &client.GithubRepository{
		ID:           r.clusterID,
		Organisation: r.declaration.Github.Organisation,
		Repository:   r.declaration.Github.Repository,
		FullName:     fullName,
		GitURL:       fmt.Sprintf("%s:%s", constant.DefaultGithubHost, fullName),
		DeployKey:    deployKey,
	}

	r.log.Info("Updating GitHub state")

	err = r.githubState.SaveGithubRepository(repo)
	if err != nil {
		return fmt.Errorf("saving github repository: %w", err)
	}
	return nil
}

func New(logger logger.Logger, flags cmdflags.Flags) (SSHKeyRotater, error) {
	o, err := okctlinit.InitializeOkctl()
	if err != nil {
		return SSHKeyRotater{}, fmt.Errorf("initializing: %w", err)
	}

	state := o.StateHandlers(o.StateNodes())

	githubDeployKeyGetter, err := upgradeGithub.New(context.Background(), o.CredentialsProvider.Github())
	if err != nil {
		return SSHKeyRotater{}, fmt.Errorf("creating github deploy key client: %w", err)
	}

	parameterService := core.NewParameterService(
		awsProvider.NewParameterCloudProvider(o.CloudProvider),
	)

	githubClient, err := github.New(context.Background(), o.CredentialsProvider.Github())
	if err != nil {
		return SSHKeyRotater{}, fmt.Errorf("creating github client: %w", err)
	}

	githubService := clientCore.NewGithubService(
		parameterService,
		*githubClient,
		state.Github,
	)

	return SSHKeyRotater{
		log:                   logger,
		flags:                 flags,
		declaration:           o.Declaration,
		clusterID:             okctlinit.GetClusterID(o),
		githubState:           state.Github,
		parameterService:      parameterService,
		githubDeployKeyGetter: githubDeployKeyGetter,
		githubClient:          githubClient,
		githubService:         githubService,
	}, nil
}
