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

func ParsePCOperatorImages(valuesFile string) (OperatorMap, OperatorMap) {
	const minMatchLength = 3
	imageRegex := regexp.MustCompile(`repository:\s*([^\s]+)[\s\S]+?version:\s*([^\s]+)[\s\S]+?`)
	matches := imageRegex.FindAllStringSubmatch(valuesFile, -1)
	operatorPcoImages := make(OperatorMap)
	operatorOtherImages := make(OperatorMap)
	for _, match := range matches {
		if len(match) < minMatchLength {
			continue
		}
		repo, version := match[1], match[2]
		switch {
		case strings.Contains(repo, "registry.redhat.io/rhtas/policy-controller-rhel9"):
			operatorPcoImages["policy-controller-image"] = fmt.Sprintf("%s@%s", repo, version)
		case strings.Contains(repo, "registry.redhat.io/openshift4/ose-cli"):
			operatorOtherImages["ose-cli-image"] = fmt.Sprintf("%s@sha256:%s", repo, version)
		}
	}
	return operatorPcoImages, operatorOtherImages
}

func ParseMVOperatorImages(valuesFile string) OperatorMap {
	operatorMvoImages := make(OperatorMap)

	const reFirstCapture = 1

	helpRe := regexp.MustCompile(`(?s)-{1,2}model-transparency-cli-image\b.*?\(default\s+"([^"]+)"\)`)
	m := helpRe.FindStringSubmatch(valuesFile)
	if len(m) > reFirstCapture {
		img := strings.TrimSpace(m[reFirstCapture])
		if img != "" {
			operatorMvoImages["model-transparency-image"] = img
			return operatorMvoImages
		}
	}

	return operatorMvoImages
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

func ExtractHash(image string) string {
	return image[len(image)-64:]
}
