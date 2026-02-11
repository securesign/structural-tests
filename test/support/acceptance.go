package support

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type OperatorMap map[string]string

func ParseSnapshotData() (SnapshotData, error) {
	snapshotFileName := GetEnv(EnvReleasesSnapshotFile)
	if snapshotFileName == "" {
		return SnapshotData{}, fmt.Errorf("snapshot file name must be set. Use %s env variable for that", EnvReleasesSnapshotFile)
	}
	content, err := GetFileContent(snapshotFileName)
	if err != nil {
		return SnapshotData{}, err
	}
	var snapshotData SnapshotData
	err = json.Unmarshal(content, &snapshotData)
	if err != nil {
		return SnapshotData{}, fmt.Errorf("failed to parse snapshot file: %w", err)
	}
	return snapshotData, nil
}

func ParseOperatorImages(helpContent string) (OperatorMap, OperatorMap) {
	const minimumValidMatches = 3
	re := regexp.MustCompile(`-([\w-]+image)\s+string[^"]+default "([^"]+)"`)
	matches := re.FindAllStringSubmatch(helpContent, -1)
	operatorTasImages := make(OperatorMap)
	operatorOtherImages := make(OperatorMap)
	for _, match := range matches {
		if len(match) >= minimumValidMatches {
			key := match[1]
			value := match[2]
			if slices.Contains(OtherOperatorImageKeys(), key) {
				operatorOtherImages[key] = value
				continue
			}
			operatorTasImages[key] = value
		}
	}
	return operatorTasImages, operatorOtherImages
}

func MapAnsibleImages(ansibleDefinitionFileContent []byte) (AnsibleMap, error) {
	var ansibleImages AnsibleMap
	err := yaml.Unmarshal(ansibleDefinitionFileContent, &ansibleImages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ansible images file: %w", err)
	}
	return ansibleImages, nil
}

func ConvertAnsibleImageKey(ansibleImageKey string) string {
	if !strings.HasPrefix(ansibleImageKey, "tas_single_node_") {
		return ansibleImageKey
	}
	result := strings.ReplaceAll(strings.TrimPrefix(ansibleImageKey, "tas_single_node_"), "_", "-")
	return result
}

func ExtractHashes(images []string) []string {
	result := make([]string, len(images))
	for i, image := range images {
		result[i] = ExtractHash(image)
	}
	return result
}

const sha256DigestLen = 64

func ExtractHash(image string) string {
	const sha256Prefix = "@sha256:"
	i := strings.LastIndex(image, sha256Prefix)
	if i < 0 {
		return ""
	}
	start := i + len(sha256Prefix)
	if start+sha256DigestLen > len(image) {
		return ""
	}
	return image[start : start+sha256DigestLen]
}
