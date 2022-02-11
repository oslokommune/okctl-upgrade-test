// Package github provides a client for interacting with the Github API
package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v32/github"
	githubAuth "github.com/oslokommune/okctl/pkg/credentials/github"
	"golang.org/x/oauth2"
)

// ErrNotFound means something was not found
var ErrNotFound = errors.New("not found")

// Githuber invokes the github API
type Githuber interface {
	GetDeployKeys(org, repository, deployKeyName string) ([]*Key, error)
}

// Github contains the state for interacting with the github API
type Github struct {
	Ctx    context.Context
	Client *github.Client
}

func (g *Github) GetDeployKeys(org, repository, deployKeyName string) ([]*Key, error) {
	allKeys, err := g.ListDeployKey(org, repository)
	if err != nil {
		return nil, fmt.Errorf("getting deploy key: %w", err)
	}

	var keysWithName []*Key

	for _, key := range allKeys {
		if key.GetTitle() == deployKeyName {
			keysWithName = append(keysWithName, key)
		}
	}

	if len(keysWithName) == 0 {
		return nil, ErrNotFound
	}

	return keysWithName, nil
}

func (g *Github) ListDeployKey(org, repository string) ([]*Key, error) {
	opts := &github.ListOptions{
		Page:    0,
		PerPage: 100,
	}

	var allKeys []*Key

	for {
		keys, response, err := g.Client.Repositories.ListKeys(g.Ctx, org, repository, opts)
		if err != nil {
			return nil, fmt.Errorf("listing deploy keys: %w", err)
		}

		allKeys = append(allKeys, keys...)

		if response.NextPage == 0 {
			break
		}

		opts.Page = response.NextPage
	}

	return allKeys, nil
}

// DeleteDeployKey removes a read-only deploy key
func (g *Github) DeleteDeployKey(org, repository string, identifier int64) error {
	_, err := g.Client.Repositories.DeleteKey(g.Ctx, org, repository, identifier)
	if err != nil {
		return fmt.Errorf("deleting deploy key: %w", err)
	}

	return nil
}

// Ensure that Github implements Githuber
var _ Githuber = &Github{}

// Key shadows github.Key
type Key = github.Key

// New returns an initialised github API client
func New(ctx context.Context, auth githubAuth.Authenticator) (*Github, error) {
	credentials, err := auth.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get github credentials: %w", err)
	}

	client := github.NewClient(
		oauth2.NewClient(ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: credentials.AccessToken,
				},
			),
		),
	)

	return &Github{
		Ctx:    ctx,
		Client: client,
	}, nil
}
