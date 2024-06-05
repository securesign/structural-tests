package support

type Snapshot struct {
	Day                       string `json:"day"`
	Time                      string `json:"time"`
	CertificateTransparencyGo struct {
		SnapshotName              string `json:"snapshot_name"`
		CertificateTransparencyGo string `json:"certificate-transparency-go"`
	} `json:"certificate-transparency-go"`
	Cli struct {
		SnapshotName   string `json:"snapshot_name"`
		ClientServerCg string `json:"client-server-cg"`
		ClientServerRe string `json:"client-server-re"`
		Cosign         string `json:"cosign"`
		Gitsign        string `json:"gitsign"`
	} `json:"cli"`
	FbcV413 struct {
		SnapshotName string `json:"snapshot_name"`
		FbcV413      string `json:"fbc-v4-13"`
	} `json:"fbc-v4-13"`
	FbcV414 struct {
		SnapshotName string `json:"snapshot_name"`
		FbcV414      string `json:"fbc-v4-14"`
	} `json:"fbc-v4-14"`
	FbcV415 struct {
		SnapshotName string `json:"snapshot_name"`
		FbcV415      string `json:"fbc-v4-15"`
	} `json:"fbc-v4-15"`
	Fulcio struct {
		SnapshotName string `json:"snapshot_name"`
		FulcioServer string `json:"fulcio-server"`
	} `json:"fulcio"`
	Operator struct {
		SnapshotName        string `json:"snapshot_name"`
		RhtasOperator       string `json:"rhtas-operator"`
		RhtasOperatorBundle string `json:"rhtas-operator-bundle"`
	} `json:"operator"`
	Rekor struct {
		SnapshotName  string `json:"snapshot_name"`
		BackfillRedis string `json:"backfill-redis"`
		RekorCli      string `json:"rekor-cli"`
		RekorServer   string `json:"rekor-server"`
	} `json:"rekor"`
	RekorSearchUI struct {
		SnapshotName  string `json:"snapshot_name"`
		RekorSearchUI string `json:"rekor-search-ui"`
	} `json:"rekor-search-ui"`
	Scaffold struct {
		SnapshotName       string `json:"snapshot_name"`
		Createctconfig     string `json:"createctconfig"`
		CtlogManagectroots string `json:"ctlog-managectroots"`
		FulcioCreatecerts  string `json:"fulcio-createcerts"`
		TrillianCreatedb   string `json:"trillian-createdb"`
		TrillianCreatetree string `json:"trillian-createtree"`
		TufServer          string `json:"tuf-server"`
	} `json:"scaffold"`
	SegmentBackupJob struct {
		SnapshotName     string `json:"snapshot_name"`
		SegmentBackupJob string `json:"segment-backup-job"`
	} `json:"segment-backup-job"`
	Trillian struct {
		SnapshotName string `json:"snapshot_name"`
		Database     string `json:"database"`
		Logserver    string `json:"logserver"`
		Logsigner    string `json:"logsigner"`
		Redis        string `json:"redis"`
	} `json:"trillian"`
}
