package support

import (
	"encoding/json"
	"regexp"
)

type OperatorMap map[string]string

func ParseSnapshotImages() (SnapshotMap, error) {
	content, err := GetFileContent(GetEnvOrDefault(EnvReleasesSnapshotFile, DefaultReleasesSnapshotFile))
	if err != nil {
		return nil, err
	}
	var snapshotImages SnapshotMap
	err = json.Unmarshal([]byte(content), &snapshotImages)
	return snapshotImages, err
}

func ParseOperatorImages(helpContent string) OperatorMap {
	re := regexp.MustCompile(`-([\w-]+image)\s+string[^"]+default "([^"]+)"`)
	matches := re.FindAllStringSubmatch(helpContent, -1)
	images := make(OperatorMap)
	for _, match := range matches {
		if len(match) > 2 {
			key := match[1]
			value := match[2]
			if key == "client-server-image" || key == "trillian-netcat-image" { // not interested in these
				continue
			}
			images[key] = value
		}
	}
	return images
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
