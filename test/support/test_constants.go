package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvAnsibleImagesFile    = "ANSIBLE"
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

func AnsibleImageKeys() []string {
	return []string{
		"tas_single_node_fulcio_server_image",
		"tas_single_node_trillian_log_server_image",
		"tas_single_node_trillian_log_signer_image",
		"tas_single_node_rekor_server_image",
		"tas_single_node_ctlog_image",
		"tas_single_node_rekor_redis_image",
		"tas_single_node_trillian_db_image",
		"tas_single_node_tuf_image",
		"tas_single_node_timestamp_authority_image",
		"tas_single_node_rekor_search_ui_image",
		"tas_single_node_createtree_image",
		"tas_single_node_client_server_image",
	}
}

func OtherAnsibleImageKeys() []string {
	return []string{
		"tas_single_node_trillian_netcat_image",
		"tas_single_node_nginx_image",
	}
}
