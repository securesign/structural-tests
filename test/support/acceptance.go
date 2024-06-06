package support

import (
	"encoding/json"
	"fmt"
	gitAuth "github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GetImageDefinitionsFromSnapshot(jsonData Snapshot) []string {
	return []string{
		jsonData.CertificateTransparencyGo.CertificateTransparencyGo,
		jsonData.Cli.ClientServerCg,
		jsonData.Cli.ClientServerRe,
		jsonData.Cli.Cosign,
		jsonData.Cli.Gitsign,
		jsonData.FbcV413.FbcV413,
		jsonData.FbcV414.FbcV414,
		jsonData.FbcV415.FbcV415,
		jsonData.Fulcio.FulcioServer,
		jsonData.Operator.RhtasOperator,
		jsonData.Operator.RhtasOperatorBundle,
		jsonData.Rekor.BackfillRedis,
		jsonData.Rekor.RekorCli,
		jsonData.Rekor.RekorServer,
		jsonData.RekorSearchUI.RekorSearchUI,
		jsonData.Scaffold.Createctconfig,
		jsonData.Scaffold.CtlogManagectroots,
		jsonData.Scaffold.FulcioCreatecerts,
		jsonData.Scaffold.TrillianCreatedb,
		jsonData.Scaffold.TrillianCreatetree,
		jsonData.Scaffold.TufServer,
		jsonData.SegmentBackupJob.SegmentBackupJob,
		jsonData.Trillian.Database,
		jsonData.Trillian.Logserver,
		jsonData.Trillian.Logsigner,
		jsonData.Trillian.Redis,
	}
}

func GetReleasesProjectPath() (string, error) {
	localReleasesPath := GetEnv(EnvLocalReleasesProjectPath)
	if localReleasesPath != "" {
		localReleasesPath := localPathCleanup(localReleasesPath)
		log.Printf("Using local folder %s\n", localReleasesPath)
		return localReleasesPath, nil
	} else {
		releasesBranch := GetEnvOrDefault(EnvReleasesRepoBranch, ReleasesRepoDefBranch)
		githubUsername := GetEnv(EnvTestGithubUser)
		githubToken := GetEnvOrDefaultSecret(EnvTestGithubToken, "")
		releasesPath, _, err := GitCloneWithAuth(ReleasesRepo, releasesBranch,
			&gitAuth.BasicAuth{
				Username: githubUsername,
				Password: githubToken,
			})

		return releasesPath, err
	}
}

func GetReleasesSnapshotFilePath() (string, bool) {
	localReleasesPath := GetEnv(EnvLocalReleasesProjectPath)
	snapshotFileFolder := GetEnvOrDefault(EnvReleasesSnapshotFolder, ReleasesSnapshotDefFolder)
	if localReleasesPath != "" {
		localReleasesPath := localPathCleanup(localReleasesPath)
		snapshotFile := filepath.Join(localReleasesPath, snapshotFileFolder, "snapshot.json")
		return snapshotFile, true
	} else {
		releasesBranch := GetEnvOrDefault(EnvReleasesRepoBranch, ReleasesRepoDefBranch)
		snapshotFile := fmt.Sprintf(ReleasesSnapshotFile, releasesBranch, snapshotFileFolder)
		return snapshotFile, false
	}
}

func localPathCleanup(origPath string) string {
	finalPath := origPath
	if !filepath.IsAbs(origPath) {
		// not ideal solution
		// want to have path relative to the project directory
		// without test/acceptance-tests
		finalPath = filepath.Join("..", "..", origPath)
	}
	return filepath.Clean(finalPath)
}

func GetCorrespondingSnapshotImage(operatorImageKey string, snapshotJsonData Snapshot) string {
	switch operatorImageKey {
	case "backfill-redis-image":
		return snapshotJsonData.Rekor.BackfillRedis
	case "client-server-cg-image":
		return snapshotJsonData.Cli.ClientServerCg
	case "client-server-re-image":
		return snapshotJsonData.Cli.ClientServerRe
	case "ctlog-image":
		return snapshotJsonData.CertificateTransparencyGo.CertificateTransparencyGo
	case "fulcio-server-image":
		return snapshotJsonData.Fulcio.FulcioServer
	case "rekor-redis-image":
		return snapshotJsonData.Trillian.Redis
	case "rekor-search-ui-image":
		return snapshotJsonData.RekorSearchUI.RekorSearchUI
	case "rekor-server-image":
		return snapshotJsonData.Rekor.RekorServer
	case "segment-backup-job-image":
		return snapshotJsonData.SegmentBackupJob.SegmentBackupJob
	case "trillian-db-image":
		return snapshotJsonData.Trillian.Database
	case "trillian-log-server-image":
		return snapshotJsonData.Trillian.Logserver
	case "trillian-log-signer-image":
		return snapshotJsonData.Trillian.Logsigner
	case "tuf-image":
		return snapshotJsonData.Scaffold.TufServer
	default:
		return ""
	}
}

func ParseOperatorImages(helpContent string) map[string]string {
	re := regexp.MustCompile(`-(\S+image)\s+string[^"]+default "([^"]+)"`)
	matches := re.FindAllStringSubmatch(helpContent, -1)

	imageMap := make(map[string]string)
	for _, match := range matches {
		if len(match) > 2 {
			key := match[1]
			value := match[2]
			imageMap[key] = value
		}
	}
	return imageMap
}

func ValidateJson(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var js interface{}
	return json.Unmarshal(content, &js)
}

func ValidateYaml(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var yml interface{}
	return yaml.Unmarshal(content, &yml)
}

func ImageHashesIdentical(imageA, imageB string) bool {
	var shaA, shaB string
	if len(imageA) >= 64 {
		shaA = imageA[len(imageA)-64:]
	} else {
		shaA = imageA
	}
	if len(imageB) >= 64 {
		shaB = imageB[len(imageB)-64:]
	} else {
		shaB = imageB
	}
	return shaA == shaB
}

func ValidateAllYamlAndJsonFiles(directory string) error {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".json") {
				validationError := ValidateJson(path)
				if validationError != nil {
					log.Printf("%s: %s", path, validationError.Error())
				}
			} else if strings.HasSuffix(info.Name(), ".yaml") {
				validationError := ValidateYaml(path)
				if validationError != nil {
					log.Printf("%s: %s", path, validationError.Error())
				}
			}
		}
		return nil
	})
	return err
}
