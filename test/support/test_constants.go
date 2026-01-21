package support

const (
	EnvReleasesSnapshotFile = "SNAPSHOT"
	EnvAnsibleImagesFile    = "ANSIBLE"
	EnvRepositoriesFile     = "REPOSITORIES"
	EnvTestGithubToken      = "TEST_GITHUB_TOKEN" // #nosec G101
	EnvVersion              = "VERSION"

	PolicyControllerOperatorImageKey       = "policy-controller-operator-image"
	PolicyControllerOperatorBundleImageKey = "policy-controller-operator-bundle-image"
	ModelValidationOperatorImageKey        = "model-validation-operator-image"
	ModelValidationOperatorBundleImageKey  = "model-validation-operator-bundle-image"
	OperatorImageKey                       = "rhtas-operator-image"
	OperatorBundleImageKey                 = "rhtas-operator-bundle-image"
	AnsibleCollectionKey                   = "artifact-signer-ansible.collection.url"

	AnsibleCollectionSnapshotFile                           = "roles/tas_single_node/defaults/main.yml"
	AnsibleArtifactsURL                                     = "https://api.github.com/repos/securesign/artifact-signer-ansible/actions/artifacts"
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

func MandatoryTasOperatorImageKeys() []string {
	return []string{
		"trillian-log-server-image",
		"trillian-log-signer-image",
		"trillian-db-image",
		"createtree-image",

		"fulcio-server-image",

		"rekor-redis-image",
		"rekor-search-ui-image",
		"rekor-server-image",
		"rekor-monitor-image",
		"backfill-redis-image",

		"tuf-image",

		"ctlog-image",
		"ctlog-monitor-image",

		"client-server-image",

		"segment-backup-job-image",

		"timestamp-authority-image",
	}
}

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
		"model-transparency-image",
	}
}

func OtherOperatorImageKeys() []string {
	return []string{
		"trillian-netcat-image",
		"http-server-image",
	}
}

func AnsibleTasImageKeys() []string {
	return []string{
		"tas_single_node_fulcio_server_image",
		"tas_single_node_trillian_log_server_image",
		"tas_single_node_trillian_log_signer_image",
		"tas_single_node_rekor_server_image",
		"tas_single_node_rekor_monitor_image",
		"tas_single_node_ctlog_image",
		"tas_single_node_ctlog_monitor_image",
		"tas_single_node_rekor_redis_image",
		"tas_single_node_trillian_db_image",
		"tas_single_node_tuf_image",
		"tas_single_node_timestamp_authority_image",
		"tas_single_node_rekor_search_ui_image",
		"tas_single_node_createtree_image",
		"tas_single_node_client_server_image",
		"tas_single_node_backfill_redis_image",
	}
}

func AnsibleOtherImageKeys() []string {
	return []string{
		"tas_single_node_http_server_image",
		"tas_single_node_trillian_netcat_image",
		"tas_single_node_nginx_image",
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
