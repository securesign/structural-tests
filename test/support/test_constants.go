package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvRepositoriesFile     = "REPOSITORIES"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN" // #nosec G101
	EnvVersion              = "VERSION"
	EnvTestConfig           = "TEST_CONFIG"

	PolicyControllerOperatorImageKey       = "policy-controller-operator-image"
	PolicyControllerOperatorBundleImageKey = "policy-controller-operator-bundle-image"
	ModelValidationOperatorImageKey        = "model-validation-operator-image"
	ModelValidationOperatorBundleImageKey  = "model-validation-operator-bundle-image"
	OperatorImageKey                       = "rhtas-operator-image"
	OperatorBundleImageKey                 = "rhtas-operator-bundle-image"
	AnsibleCollectionImageKey              = "artifact-signer-ansible.collection.image"
	AnsibleCollectionPathInImage           = "/releases"

	AnsibleCollectionSnapshotFile                           = "roles/tas_single_node/defaults/main.yml"
	OperatorBundleClusterServiceVersionFile                 = "rhtas-operator.clusterserviceversion.yaml"
	PolicyControllerOperatorBundleClusterServiceVersionFile = "policy-controller-operator.clusterserviceversion.yaml"
	ModelValidationOperatorBundleClusterServiceVersionFile  = "model-validation-operator.clusterserviceversion.yaml"
	OperatorBundleClusterServiceVersionPath                 = "manifests/" + OperatorBundleClusterServiceVersionFile
	PolicyControllerOperatorBundleClusterServiceVersionPath = "manifests/" + PolicyControllerOperatorBundleClusterServiceVersionFile
	ModelValidationOperatorBundleClusterServiceVersionPath  = "manifests/" + ModelValidationOperatorBundleClusterServiceVersionFile

	TasImageDefinitionRegexp      = `^registry.redhat.io/rhtas/[\w/-]+@sha256:\w{64}$`
	OtherImageDefinitionRegexp    = `^(registry.redhat.io|registry.access.redhat.com)`
	SnapshotImageDefinitionRegexp = `^[\.\w/-]+@sha256:\w{64}$`

	DefaultRepositoriesFile = "testdata/repositories.json"
)

func MandatoryPcoOperatorImageKeys() []string {
	return []string{
		"policy-controller-image",
	}
}

func OtherPCOOperatorImageKeys() []string {
	return []string{
		"ose-cli-image",
	}
}

func MandatoryMvoOperatorImageKeys() []string {
	return []string{
		"model-validation-agent-image",
	}
}

type OSArchMatrix map[string][]string

func GetOSArchMatrix() OSArchMatrix {
	return map[string][]string{
		"linux":   {"amd64", "arm64", "ppc64le", "s390x"},
		"darwin":  {"amd64", "arm64"},
		"windows": {"amd64"},
	}
}

// If no value is provided, the label must exist, but can have any non-empty value.
func RequiredImageLabels() map[string]string {
	return map[string]string{
		"architecture": "x86_64",
		"build-date":   "",
		"vcs-ref":      "",
		"vcs-type":     "git",
		"vendor":       "Red Hat, Inc.",
	}
}
