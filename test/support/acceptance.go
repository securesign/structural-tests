package support

import (
	"regexp"
)

type OperatorMap map[string]string

func ParseOperatorImages(helpContent string) map[string]string {
	re := regexp.MustCompile(`-(\S+image)\s+string[^"]+default "([^"]+)"`)
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
	var result []string
	for _, image := range images {
		result = append(result, image[len(image)-64:])
	}
	return result
}
