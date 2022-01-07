package argocd

import (
	"fmt"
	"github.com/oslokommune/okctl/cmd/okctl/hooks"
	"github.com/oslokommune/okctl/pkg/api"
	"github.com/oslokommune/okctl/pkg/config/constant"
	"github.com/oslokommune/okctl/pkg/okctl"
	"github.com/spf13/cobra"
	"os"
	"path"
)

const (
	localStatePathErrFormat = "acquiring local state path: %w"
)

func initializeOkctl() (*okctl.Okctl, error) {
	o := okctl.New()
	cmd := &cobra.Command{}
	args := []string{}

	err := hooks.LoadUserData(o)(cmd, args)
	if err != nil {
		return nil, fmt.Errorf("loading user data: %w", err)
	}

	err = initializeDeclaration(o)
	if err != nil {
		return nil, fmt.Errorf("initializing declaration: %w", err)
	}

	err = o.Initialise()
	if err != nil {
		return nil, fmt.Errorf("initializing okctl: %w", err)
	}

	err = initializeState(o)
	if err != nil {
		return nil, fmt.Errorf("initializing state: %w", err)
	}

	return o, nil
}

func initializeDeclaration(o *okctl.Okctl) error {
	clusterDeclarationPath := os.Getenv(constant.EnvClusterDeclaration)
	if clusterDeclarationPath == "" {
		return fmt.Errorf("missing required %s environment variable", constant.EnvClusterDeclaration)
	}

	declaration, err := readClusterDeclaration(clusterDeclarationPath)
	if err != nil {
		return fmt.Errorf("reading cluster declaration: %w", err)
	}

	err = declaration.Validate()
	if err != nil {
		return fmt.Errorf("validating cluster declaration: %w", err)
	}

	o.Declaration = declaration

	return nil
}

func initializeState(o *okctl.Okctl) error {
	localStateDBPath, err := getLocalStatePath(o)
	if err != nil {
		return fmt.Errorf(localStatePathErrFormat, err)
	}

	_, err = os.Stat(localStateDBPath)
	if err != nil {
		return err
	}

	o.DB.SetDatabaseFilePath(localStateDBPath)
	o.DB.SetWritable(true)

	return nil
}

func getLocalStatePath(o *okctl.Okctl) (string, error) {
	dataDir, err := o.GetUserDataDir()
	if err != nil {
		return "", fmt.Errorf("acquiring user data dir: %w", err)
	}

	dir := path.Join(dataDir, "localState", o.Declaration.Metadata.Name)

	err = o.FileSystem.MkdirAll(dir, 0o700)
	if err != nil {
		return "", fmt.Errorf("creating temp state folder: %w", err)
	}

	return path.Join(dir, constant.DefaultStormDBName), nil
}

func getClusterID(o *okctl.Okctl) api.ID {
	return api.ID{
		Region:       o.Declaration.Metadata.Region,
		AWSAccountID: o.Declaration.Metadata.AccountID,
		ClusterName:  o.Declaration.Metadata.Name,
	}
}
