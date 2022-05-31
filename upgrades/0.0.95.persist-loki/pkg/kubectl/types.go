// Package kubectl exposes a simplified API for operations done with kubectl
package kubectl

type secret struct {
	Data map[string]interface{} `json:"data"`
}

type pvcSpec struct {
	VolumeName string `json:"volumeName"`
}

type metadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type pvc struct {
	Metadata metadata `json:"metadata"`
	Spec     pvcSpec  `json:"spec"`
}

type pv struct {
	Metadata metadata `json:"metadata"`
}

type iteratorItem struct {
	Metadata metadata `json:"metadata"`
}

type iterator struct {
	Items []iteratorItem `json:"items"`
}
