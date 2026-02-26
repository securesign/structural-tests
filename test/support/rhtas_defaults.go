package support

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// MergeRhtasConfig overlays fileContent on top of baseDefaults. Keys present in fileContent
// override baseDefaults; keys only in baseDefaults (e.g. operatorOtherImageKeys) are preserved.
// Use this when TEST_CONFIG points to a partial config file so embedded defaults fill in missing keys.
func MergeRhtasConfig(baseDefaults, fileContent []byte) ([]byte, error) {
	if len(fileContent) == 0 {
		return baseDefaults, nil
	}
	var base, overlay map[string]interface{}
	if err := yaml.Unmarshal(baseDefaults, &base); err != nil {
		return nil, fmt.Errorf("parse base defaults: %w", err)
	}
	if err := yaml.Unmarshal(fileContent, &overlay); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	for k, v := range overlay {
		base[k] = v
	}
	merged, err := yaml.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("marshal merged config: %w", err)
	}
	return merged, nil
}

// rhtasDefaults is the subset of test/acceptance/rhtas/defaults.yaml used for operator and ansible image keys.
type rhtasDefaults struct {
	OperatorImageKeys      []string `yaml:"operatorImageKeys"`
	OperatorOtherImageKeys []string `yaml:"operatorOtherImageKeys"`
	AnsibleImageKeys       []string `yaml:"ansibleImageKeys"`
	AnsibleOtherImageKeys  []string `yaml:"ansibleOtherImageKeys"`
}

// GetOperatorImageKeysFromConfig returns the TAS operator image key list from
// rhtas defaults.yaml or TEST_CONFIG. Returns an error if config is missing or the key is empty.
func GetOperatorImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return nil, errors.New("defaults config is required for operatorImageKeys")
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.OperatorImageKeys) == 0 {
		return nil, errors.New("operatorImageKeys is missing or empty in config")
	}
	return parsed.OperatorImageKeys, nil
}

// GetOperatorOtherImageKeysFromConfig returns the other (non-TAS) operator image key list from
// rhtas defaults.yaml or TEST_CONFIG. Returns an error if config is missing or the key is empty.
func GetOperatorOtherImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return nil, errors.New("defaults config is required for operatorOtherImageKeys")
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.OperatorOtherImageKeys) == 0 {
		return nil, errors.New("operatorOtherImageKeys is missing or empty in config")
	}
	return parsed.OperatorOtherImageKeys, nil
}

// GetAnsibleImageKeysFromConfig returns both ansible TAS and other image key lists from
// rhtas defaults.yaml or TEST_CONFIG. Returns an error if config is missing or a key is empty.
func GetAnsibleImageKeysFromConfig(defaultsYaml []byte) ([]string, []string, error) {
	if len(defaultsYaml) == 0 {
		return nil, nil, errors.New("defaults config is required for ansible image keys")
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.AnsibleImageKeys) == 0 {
		return nil, nil, errors.New("ansibleImageKeys is missing or empty in config")
	}
	if len(parsed.AnsibleOtherImageKeys) == 0 {
		return nil, nil, errors.New("ansibleOtherImageKeys is missing or empty in config")
	}
	return parsed.AnsibleImageKeys, parsed.AnsibleOtherImageKeys, nil
}
