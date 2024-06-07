package support

const (
	DefaultReleasesSnapshotFile = "https://raw.githubusercontent.com/securesign/releases/main/1.0.1/snapshot.json"

	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN"

	ImageDefinitionRegexp = `\S\w+@sha256:\w{64}$`
	OperatorImageKey      = "rhtas-operator"
)
