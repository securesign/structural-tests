package support

import (
	"encoding/json"
	"strings"
)

type SnapshotMap map[string]string

func (data *SnapshotMap) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*data = make(map[string]string)
	extractImages(raw, *data)
	return nil
}

func extractImages(data map[string]interface{}, images map[string]string) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if isImageDefinition(v) {
				images[key] = v
			}
		case map[string]interface{}:
			extractImages(v, images)
		}
	}
}

func isImageDefinition(snapshotValue string) bool {
	return len(snapshotValue) > 30 &&
		strings.Count(snapshotValue, ".") >= 1 &&
		strings.Count(snapshotValue, "/") >= 2
}
