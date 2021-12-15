package grafana

const jsonPatchOperationReplace = "replace"

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
