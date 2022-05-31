// Package kubectl exposes a simplified API for operations done with kubectl
package kubectl

type metadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type iteratorItem struct {
	Metadata metadata `json:"metadata"`
}

type iterator struct {
	Items []iteratorItem `json:"items"`
}
