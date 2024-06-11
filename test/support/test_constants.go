package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN"

	OperatorImageKey       = "rhtas-operator"
	OperatorBundleImageKey = "rhtas-operator-bundle"

	DefaultReleasesSnapshotFile             = "https://raw.githubusercontent.com/securesign/releases/main/1.0.1/snapshot.json"
	OperatorBundleClusterserviceversionFile = "manifests/rhtas-operator.clusterserviceversion.yaml"

	ImageDefinitionRegexp = `\S\w+@sha256:\w{64}$`
)
