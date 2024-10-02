package support

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type AnsibleMap map[string]string

func (data *AnsibleMap) UnmarshalYAML(value *yaml.Node) error {
	*data = make(map[string]string)

	var rawMap map[string]interface{}
	if err := value.Decode(&rawMap); err != nil {
		return fmt.Errorf("error while parsing yaml file: %w", err)
	}

	for key, val := range rawMap {
		if strings.HasSuffix(key, "image") {
			if strVal, ok := val.(string); ok {
				(*data)[key] = strVal
			}
		}
	}

	return nil
}
