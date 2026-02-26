package support

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// MergeRhtasConfig overlays fileContent on top of baseDefaults. Keys present in fileContent
// override baseDefaults; keys only in baseDefaults (e.g. otherOperatorImageKeys) are preserved.
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
	OtherOperatorImageKeys        []string `yaml:"otherOperatorImageKeys"`
	MandatoryTasOperatorImageKeys []string `yaml:"mandatoryTasOperatorImageKeys"`
	AnsibleTasImageKeys           []string `yaml:"ansibleTasImageKeys"`
	AnsibleOtherImageKeys         []string `yaml:"ansibleOtherImageKeys"`
}

// GetMandatoryTasOperatorImageKeysFromConfig returns the mandatory TAS operator image key list from
// rhtas defaults.yaml or TEST_CONFIG. Returns an error if config is missing or the key is empty.
func GetMandatoryTasOperatorImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return nil, fmt.Errorf("defaults config is required for mandatoryTasOperatorImageKeys")
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.MandatoryTasOperatorImageKeys) == 0 {
		return nil, fmt.Errorf("mandatoryTasOperatorImageKeys is missing or empty in config")
	}
	return parsed.MandatoryTasOperatorImageKeys, nil
}

// GetOtherOperatorImageKeysFromConfig returns the other (non-TAS) operator image key list from
// rhtas defaults.yaml or TEST_CONFIG. Returns an error if config is missing or the key is empty.
func GetOtherOperatorImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return nil, fmt.Errorf("defaults config is required for otherOperatorImageKeys")
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.OtherOperatorImageKeys) == 0 {
		return nil, fmt.Errorf("otherOperatorImageKeys is missing or empty in config")
	}
	return parsed.OtherOperatorImageKeys, nil
}

// GetAnsibleImageKeysFromConfig returns both ansible TAS and other image key lists from
// rhtas defaults.yaml or TEST_CONFIG. Returns an error if config is missing or a key is empty.
func GetAnsibleImageKeysFromConfig(defaultsYaml []byte) ([]string, []string, error) {
	if len(defaultsYaml) == 0 {
		return nil, nil, fmt.Errorf("defaults config is required for ansible image keys")
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.AnsibleTasImageKeys) == 0 {
		return nil, nil, fmt.Errorf("ansibleTasImageKeys is missing or empty in config")
	}
	if len(parsed.AnsibleOtherImageKeys) == 0 {
		return nil, nil, fmt.Errorf("ansibleOtherImageKeys is missing or empty in config")
	}
	return parsed.AnsibleTasImageKeys, parsed.AnsibleOtherImageKeys, nil
}
