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
		"tuf-image",
		"trillian-log-server-image",
		"trillian-log-signer-image",
		"trillian-db-image",
		"rekor-redis-image",
		"rekor-search-ui-image",
		"rekor-server-image",
		"fulcio-server-image",
		"client-server-cg-image",
		"client-server-re-image",
		"ctlog-image",
		"backfill-redis-image",
		"segment-backup-job-image",
	}
}

func OtherOperatorImageKeys() []string {
	return []string{
		"client-server-image",
		"trillian-netcat-image",
	}
}
