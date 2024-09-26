package support

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
)

type OperatorMap map[string]string

func ParseSnapshotImages() (SnapshotMap, error) {
	snapshotFileName := GetEnv(EnvReleasesSnapshotFile)
	if snapshotFileName == "" {
		return nil, errors.New(fmt.Sprintf("snapshot file name must be set. Use %s env variable for that", EnvReleasesSnapshotFile))
	}
	content, err := GetFileContent(snapshotFileName)
	if err != nil {
		return nil, err
	}
	var snapshotImages SnapshotMap
	err = json.Unmarshal([]byte(content), &snapshotImages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot file: %w", err)
	}
	return snapshotImages, nil
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

func ParseAnsibleImages() (AnsibleMap, error) {
	ansibleFileName := GetEnv(EnvAnsibleImagesFile)
	if ansibleFileName == "" {
		return nil, errors.New(fmt.Sprintf("ansible images file name must be set. Use %s env variable for that", EnvAnsibleImagesFile))
	}
	content, err := GetFileContent(ansibleFileName)
	if err != nil {
		return nil, err
	}
	var ansibleImages AnsibleMap
	err = yaml.Unmarshal([]byte(content), &ansibleImages)
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

func ExtractHash(image string) string {
	return image[len(image)-64:]
}
