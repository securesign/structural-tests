package support

const (
	ReleasesRepo              = "https://github.com/securesign/releases.git"
	ReleasesSnapshotFile      = "https://raw.githubusercontent.com/securesign/releases/%s/%s/snapshot.json"
	ReleasesRepoDefBranch     = "main"
	ReleasesSnapshotDefFolder = "1.0.1"

	EnvTestGithubUser         = "TEST_GITHUB_USER"
	EnvTestGithubToken        = "TEST_GITHUB_TOKEN"
	EnvReleasesRepoBranch     = "RELEASES_BRANCH"
	EnvReleasesSnapshotFolder = "RELEASES_SNAPSHOT_FOLDER"
)
