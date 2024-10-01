package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvRepositoriesFile     = "REPOSITORIES"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN" // #nosec G101

	OperatorImageKey       = "rhtas-operator-image"
	OperatorBundleImageKey = "rhtas-operator-bundle-image"

	OperatorBundleClusterServiceVersionFile = "manifests/rhtas-operator.clusterserviceversion.yaml"

	OperatorTasImageDefinitionRegexp   = `^registry.redhat.io/rhtas/[\w/-]+@sha256:\w{64}$`
	OtherOperatorImageDefinitionRegexp = `^(registry.redhat.io|registry.access.redhat.com)`
	SnapshotImageDefinitionRegexp      = `^[\.\w/-]+@sha256:\w{64}$`

	DefaultRepositoriesFile = "testdata/repositories.json"
)

func MandatoryTasOperatorImageKeys() []string {
	return []string{
		"trillian-log-server-image",
		"trillian-log-signer-image",
		"trillian-db-image",

		"fulcio-server-image",

		"rekor-redis-image",
		"rekor-search-ui-image",
		"rekor-server-image",
		"backfill-redis-image",

		"tuf-image",

		"ctlog-image",

		"client-server-image",

		"segment-backup-job-image",

		"timestamp-authority-image",
	}
}

func OtherOperatorImageKeys() []string {
	return []string{
		"trillian-netcat-image",
		"http-server-image",
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
