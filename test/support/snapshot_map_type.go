package support

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type SnapshotMap struct {
	Images map[string]string
	Others map[string]string
}

var imageRegexp = regexp.MustCompile(`^(fbc-[\w-]+|[\w-]+-image)$`)

func (data *SnapshotMap) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("error while parsing json file: %w", err)
	}
	*data = SnapshotMap{
		make(map[string]string),
		make(map[string]string),
	}
	extractImages(raw, *data)
	return nil
}

func extractImages(data map[string]interface{}, images SnapshotMap) {
	for key, value := range data {
		switch valueType := value.(type) {
		case string:
			if isImageDefinition(key) {
				images.Images[key] = valueType
			}
		case map[string]interface{}:
			if key == "artifact-signer-ansible" {
				if collection, ok := value.(map[string]interface{})["collection"].(map[string]interface{}); ok {
					if url, ok := collection["url"].(string); ok {
						images.Others[AnsibleCollectionKey] = url
					}
				}
			} else {
				extractImages(valueType, images)
			}
		}
	}
}

func isImageDefinition(snapshotKey string) bool {
	return imageRegexp.MatchString(snapshotKey)
}
