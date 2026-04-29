package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvRepositoriesFile     = "REPOSITORIES"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN" // #nosec G101
	EnvVersion              = "VERSION"
	EnvTestConfig           = "TEST_CONFIG"

	OperatorImageKey             = "rhtas-operator-image"
	OperatorBundleImageKey       = "rhtas-operator-bundle-image"
	AnsibleCollectionImageKey    = "artifact-signer-ansible.collection.image"
	AnsibleCollectionPathInImage = "/releases"

	AnsibleCollectionSnapshotFile = "roles/tas_single_node/defaults/main.yml"

	TasImageDefinitionRegexp      = `^registry.redhat.io/rhtas/[\w/-]+@sha256:\w{64}$`
	OtherImageDefinitionRegexp    = `^(registry.redhat.io|registry.access.redhat.com)`
	SnapshotImageDefinitionRegexp = `^[\.\w/-]+@sha256:\w{64}$`

	DefaultRepositoriesFile = "testdata/repositories.json"
)

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
