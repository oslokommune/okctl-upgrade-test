package kubectl

const (
	defaultOkctlConfigDirName      = ".okctl"
	defaultOkctlBinariesDirName    = "binaries"
	defaultBinaryName              = "kubectl"
	defaultArch                    = "amd64"
	defaultAWSIAMAuthenticatorName = "aws-iam-authenticator"
	envPartsLength                 = 2

	// Kubernetes kinds
	persistentVolumeClaimResourceKind = "persistentvolumeclaim"
	persistentVolumeResourceKind      = "persistentvolume"
	secretResourceKind                = "secret"
	podResourceKind                   = "pod"
	statefulSetResourceKind           = "statefulset"
	serviceAccountResourceKind        = "serviceaccount"
)
