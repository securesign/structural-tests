package support

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GetImageDefinitionsFromConstants(constantFileContent string) ([]string, error) {
	var imageDefinitions []string
	constPattern := regexp.MustCompile(`\s+(\w+)\s*=\s*"([^"]+)"`)

	matches := constPattern.FindAllStringSubmatch(constantFileContent, -1)
	for _, match := range matches {
		if len(match) == 3 {
			imageDefinitions = append(imageDefinitions, match[2])
		}
	}

	return imageDefinitions, nil
}

func GetImageDefinitionsFromJson(jsonData map[string]interface{}) []string {
	var images []string
	for _, value := range jsonData {
		switch v := value.(type) {
		case map[string]interface{}:
			images = append(images, GetImageDefinitionsFromJson(v)...)
		case string:
			if strings.HasPrefix(v, "quay.io") ||
				strings.HasPrefix(v, "registry.redhat.io") ||
				strings.Contains(v, "@sha256:") {
				images = append(images, v)
			}
		}
	}
	return images
}

func HasValidSHA(imageDefinition string) bool {
	const pattern = `[a-f0-9]{64}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(imageDefinition)
}

func GetImageSHA(image string) (string, error) {
	parts := strings.Split(image, "@sha256:")
	if len(parts) != 2 {
		return "", errors.New("Image does not contain a SHA: " + image)
	}
	return parts[1], nil
}

func ExtractImageHashes(images []string, filteringOnly string) ([]string, error) {
	var imageSHAs []string
	for _, image := range images {
		if filteringOnly != "" && !strings.Contains(image, filteringOnly) {
			continue
		}
		sha, err := GetImageSHA(image)
		if err != nil {
			return imageSHAs, err
		}
		imageSHAs = append(imageSHAs, sha)
	}
	return imageSHAs, nil
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
