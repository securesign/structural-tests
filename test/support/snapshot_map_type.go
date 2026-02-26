package support

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type SnapshotData struct {
	Images map[string]string
	Others map[string]string
}

var (
	imageRegexp    = regexp.MustCompile(`^(fbc-[\w-]+|pco-fbc-[\w-]+|mvo-fbc-[\w-]+|[\w-]+-image)$`)
	ansibleImageRe = regexp.MustCompile(`^ansible-v1-\d+$`)
)

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
			if key == "ansible" || ansibleImageRe.MatchString(key) {
				// Image-based format: "ansible" or "ansible-v1-3": { "artifact-signer-ansible" or "artifact-signer-ansible-v1-3": "quay.io/...@sha256:..." }
				for subKey, subVal := range valueType {
					if strings.Contains(subKey, "artifact-signer-ansible") {
						if img, ok := subVal.(string); ok && img != "" {
							snapshotData.Others[AnsibleCollectionImageKey] = img
							break
						}
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
