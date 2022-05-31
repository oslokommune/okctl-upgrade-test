package resources

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type PodSpec struct {
	Containers []Container `json:"containers"`
}

type Pod struct {
	Spec PodSpec `json:"spec"`
}

type DeploymentSpec struct {
	Template Pod `json:"template"`
}

type Deployment struct {
	Spec DeploymentSpec `json:"spec"`
}
