package support

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
