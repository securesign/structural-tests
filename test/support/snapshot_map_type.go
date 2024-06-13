package support

import (
	"encoding/json"
	"regexp"
)

type SnapshotMap map[string]string

var re = regexp.MustCompile(`^[\w-]+-image$`)

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
			if isImageDefinition(key) {
				images[key] = v
			}
		case map[string]interface{}:
			extractImages(v, images)
		}
	}
}

func isImageDefinition(snapshotKey string) bool {
	return re.MatchString(snapshotKey)
}
