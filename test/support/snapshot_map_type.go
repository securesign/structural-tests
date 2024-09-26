package support

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type SnapshotMap map[string]string

var imageRegexp = regexp.MustCompile(`^(fbc-[\w-]+|[\w-]+-image)$`)

func (data *SnapshotMap) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("error while parsing json file: %w", err)
	}
	*data = make(map[string]string)
	extractImages(raw, *data)
	return nil
}

func extractImages(data map[string]interface{}, images map[string]string) {
	for key, value := range data {
		switch valueType := value.(type) {
		case string:
			if isImageDefinition(key) {
				images[key] = valueType
			}
		case map[string]interface{}:
			extractImages(valueType, images)
		}
	}
}

func isImageDefinition(snapshotKey string) bool {
	return imageRegexp.MatchString(snapshotKey)
}
