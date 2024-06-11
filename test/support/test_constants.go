package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN"
	EnvSnapshotImagesCount  = "SN_IMG_COUNT"
	EnvOperatorImagesCount  = "OP_IMG_COUNT"

	DefaultReleasesSnapshotFile = "https://raw.githubusercontent.com/securesign/releases/main/1.0.1/snapshot.json"
	ExpectedSnapshotImagesCount = 26
	ExpectedOperatorImagesCount = 13

	OperatorImageKey       = "rhtas-operator"
	OperatorBundleImageKey = "rhtas-operator-bundle"

	OperatorBundleClusterserviceversionFile = "manifests/rhtas-operator.clusterserviceversion.yaml"

	ImageDefinitionRegexp = `\S\w+@sha256:\w{64}$`
)
