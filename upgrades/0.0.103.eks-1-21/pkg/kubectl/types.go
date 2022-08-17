package kubectl

type Client interface {
	GetVersion() (Versions, error)
}

type kubectlClient struct{}
