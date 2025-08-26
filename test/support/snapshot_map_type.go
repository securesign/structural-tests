package support

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type SnapshotData struct {
	Images map[string]string
	Others map[string]string
}

var imageRegexp = regexp.MustCompile(`^(fbc-[\w-]+|pco-fbc-[\w-]+|[\w-]+-image)$`)

func (data *SnapshotData) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("error while parsing json file: %w", err)
	}
	*data = SnapshotData{
		make(map[string]string),
		make(map[string]string),
	}
	extractImages(raw, *data)
	return nil
}

func extractImages(rawData map[string]interface{}, snapshotData SnapshotData) {
	for key, value := range rawData {
		switch valueType := value.(type) {
		case string:
			if isImageDefinition(key) {
				snapshotData.Images[key] = valueType
			}
		case map[string]interface{}:
			if key == "artifact-signer-ansible" {
				if collection, ok := value.(map[string]interface{})["collection"].(map[string]interface{}); ok {
					if url, ok := collection["url"].(string); ok {
						snapshotData.Others[AnsibleCollectionKey] = url
					}
				}
			} else {
				extractImages(valueType, snapshotData)
			}
		}
	}
}

func isImageDefinition(snapshotKey string) bool {
	return imageRegexp.MatchString(snapshotKey)
}
