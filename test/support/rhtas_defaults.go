package support

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// rhtasDefaults is the subset of test/acceptance/rhtas/defaults.yaml used for operator and ansible image keys.
type rhtasDefaults struct {
	MandatoryTasOperatorImageKeys []string `yaml:"mandatoryTasOperatorImageKeys"`
	AnsibleTasImageKeys            []string `yaml:"ansibleTasImageKeys"`
	AnsibleOtherImageKeys          []string `yaml:"ansibleOtherImageKeys"`
}

// GetMandatoryTasOperatorImageKeysFromConfig returns the mandatory TAS operator image key list from
// rhtas defaults.yaml. If defaultsYaml is nil/empty or the key is missing, returns MandatoryTasOperatorImageKeys() as fallback.
func GetMandatoryTasOperatorImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return MandatoryTasOperatorImageKeys(), nil
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(parsed.MandatoryTasOperatorImageKeys) == 0 {
		return MandatoryTasOperatorImageKeys(), nil
	}
	return parsed.MandatoryTasOperatorImageKeys, nil
}

// GetAnsibleImageKeysFromConfig returns both ansible TAS and other image key lists from rhtas defaults.yaml.
// If defaultsYaml is nil/empty or a key is missing, returns AnsibleTasImageKeys() / AnsibleOtherImageKeys() as fallback.
func GetAnsibleImageKeysFromConfig(defaultsYaml []byte) ([]string, []string, error) {
	if len(defaultsYaml) == 0 {
		return AnsibleTasImageKeys(), AnsibleOtherImageKeys(), nil
	}
	var parsed rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &parsed); err != nil {
		return nil, nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	tasKeys := parsed.AnsibleTasImageKeys
	if len(tasKeys) == 0 {
		tasKeys = AnsibleTasImageKeys()
	}
	otherKeys := parsed.AnsibleOtherImageKeys
	if len(otherKeys) == 0 {
		otherKeys = AnsibleOtherImageKeys()
	}
	return tasKeys, otherKeys, nil
}
