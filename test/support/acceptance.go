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

func ParseOperatorImages(helpContent string) map[string]string {
	re := regexp.MustCompile(`-(\S+image)\s+string[^"]+default "([^"]+)"`)
	matches := re.FindAllStringSubmatch(helpContent, -1)
	imageMap := make(map[string]string)
	for _, match := range matches {
		if len(match) > 2 {
			key := match[1]
			value := match[2]
			if key == "client-server-image" || key == "trillian-netcat-image" { // not interested in these
				continue
			}
			imageMap[key] = value
		}
	}
	return imageMap
}

func ExtractHashes(imageDefinitions []string) []string {
	var result []string
	for _, image := range imageDefinitions {
		result = append(result, image[len(image)-64:])
	}
	return result
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
