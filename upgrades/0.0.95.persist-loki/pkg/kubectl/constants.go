package kubectl

const (
	// AvailabilityZoneLabelKey defines the key of the label that indicates AZ
	AvailabilityZoneLabelKey       = "failure-domain.beta.kubernetes.io/zone"
	defaultOkctlConfigDirName      = ".okctl"
	defaultOkctlBinariesDirName    = "binaries"
	defaultBinaryName              = "kubectl"
	defaultMonitoringNamespace     = "monitoring"
	defaultArch                    = "amd64"
	defaultAWSIAMAuthenticatorName = "aws-iam-authenticator"
	defaultLokiPodName             = "loki-0"
	defaultLokiConfigSecretKey     = "loki.yaml"
	envPartsLength                 = 2

	// Kubernetes kinds
	persistentVolumeClaimResourceKind = "persistentvolumeclaim"
	persistentVolumeResourceKind      = "persistentvolume"
	secretResourceKind                = "secret"
	podResourceKind                   = "pod"
	statefulSetResourceKind           = "statefulset"
	serviceAccountResourceKind        = "serviceaccount"
)
