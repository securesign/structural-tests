package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN"

	DefaultReleasesSnapshotFile = "https://raw.githubusercontent.com/securesign/releases/main/1.0.1/snapshot.json"

	OperatorImageKey       = "rhtas-operator-image"
	OperatorBundleImageKey = "rhtas-operator-bundle-image"

	OperatorBundleClusterserviceversionFile = "manifests/rhtas-operator.clusterserviceversion.yaml"

	OperatorImageDefinitionRegexp = `^registry.redhat.io/rhtas/[\w/-]+@sha256:\w{64}$`
	SnapshotImageDefinitionRegexp = `^[\.\w/-]+@sha256:\w{64}$`
)

var (
	MandatoryOperatorImageKeys = []string{
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
)
