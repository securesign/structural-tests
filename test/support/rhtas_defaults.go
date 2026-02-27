package support

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// SuiteLevelMap returns config with operator/ansible/fbc at top level.
// If the root has a "rhtas" key (package wrapper), returns that inner map; otherwise returns root.
func SuiteLevelMap(data []byte) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if raw == nil {
		return nil, errors.New("config is empty")
	}
	if rhtas, ok := raw["rhtas"]; ok {
		if m, ok := toMapAny(rhtas); ok {
			return m, nil
		}
	}
	return raw, nil
}

func toMapAny(input interface{}) (map[string]interface{}, bool) {
	if input == nil {
		return nil, false
	}
	if m, ok := input.(map[string]interface{}); ok {
		return m, true
	}
	if m, ok := input.(map[interface{}]interface{}); ok {
		out := make(map[string]interface{}, len(m))
		for k, val := range m {
			if ks, ok := k.(string); ok {
				out[ks] = val
			}
		}
		return out, true
	}
	return nil, false
}

// MergeRhtasConfig overlays fileContent on top of baseDefaults. Both may use package wrapper (rhtas:)
// or suite-level (operator, ansible, fbc). Result is always suite-level. Keys in fileContent override baseDefaults.
func MergeRhtasConfig(baseDefaults, fileContent []byte) ([]byte, error) {
	if len(fileContent) == 0 {
		return baseDefaults, nil
	}
	base, err := SuiteLevelMap(baseDefaults)
	if err != nil {
		return nil, fmt.Errorf("base defaults: %w", err)
	}
	overlay, err := SuiteLevelMap(fileContent)
	if err != nil {
		return nil, fmt.Errorf("config file: %w", err)
	}
	for key, overlayVal := range overlay {
		if key == "fbc" {
			// Deep-merge fbc so overlay does not wipe base fields (e.g. catalogPath)
			baseFbc, ok1 := toMapAny(base["fbc"])
			overlayFbc, ok2 := toMapAny(overlayVal)
			if ok1 && ok2 {
				for kk, vv := range overlayFbc {
					baseFbc[kk] = vv
				}
				base["fbc"] = baseFbc
				continue
			}
		}
		base[key] = overlayVal
	}
	merged, err := yaml.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("marshal merged config: %w", err)
	}
	return merged, nil
}

// rhtasSuites is the suite-based config: operator, ansible, fbc (no package level).
type rhtasSuites struct {
	Operator struct {
		ImageKeys      []string `yaml:"imageKeys"`
		OtherImageKeys []string `yaml:"otherImageKeys"`
	} `yaml:"operator"`
	Ansible struct {
		ImageKeys      []string `yaml:"imageKeys"`
		OtherImageKeys []string `yaml:"otherImageKeys"`
	} `yaml:"ansible"`
}

func parseSuites(defaultsYaml []byte) (rhtasSuites, error) {
	var parsed rhtasSuites
	suites, err := SuiteLevelMap(defaultsYaml)
	if err != nil {
		return parsed, err
	}
	bytes, err := yaml.Marshal(suites)
	if err != nil {
		return parsed, fmt.Errorf("marshal suites: %w", err)
	}
	if err := yaml.Unmarshal(bytes, &parsed); err != nil {
		return parsed, fmt.Errorf("parse rhtas suites: %w", err)
	}
	return parsed, nil
}

// GetOperatorImageKeysFromConfig returns the TAS operator image key list (operator.imageKeys).
func GetOperatorImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return nil, errors.New("defaults config is required for operator.imageKeys")
	}
	parsed, err := parseSuites(defaultsYaml)
	if err != nil {
		return nil, err
	}
	if len(parsed.Operator.ImageKeys) == 0 {
		return nil, errors.New("operator.imageKeys is missing or empty in config")
	}
	return parsed.Operator.ImageKeys, nil
}

// GetOperatorOtherImageKeysFromConfig returns the other operator image key list (operator.otherImageKeys).
func GetOperatorOtherImageKeysFromConfig(defaultsYaml []byte) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return nil, errors.New("defaults config is required for operator.otherImageKeys")
	}
	parsed, err := parseSuites(defaultsYaml)
	if err != nil {
		return nil, err
	}
	if len(parsed.Operator.OtherImageKeys) == 0 {
		return nil, errors.New("operator.otherImageKeys is missing or empty in config")
	}
	return parsed.Operator.OtherImageKeys, nil
}

// GetAnsibleImageKeysFromConfig returns ansible imageKeys and otherImageKeys (ansible.imageKeys, ansible.otherImageKeys).
func GetAnsibleImageKeysFromConfig(defaultsYaml []byte) ([]string, []string, error) {
	if len(defaultsYaml) == 0 {
		return nil, nil, errors.New("defaults config is required for ansible image keys")
	}
	parsed, err := parseSuites(defaultsYaml)
	if err != nil {
		return nil, nil, err
	}
	if len(parsed.Ansible.ImageKeys) == 0 {
		return nil, nil, errors.New("ansible.imageKeys is missing or empty in config")
	}
	if len(parsed.Ansible.OtherImageKeys) == 0 {
		return nil, nil, errors.New("ansible.otherImageKeys is missing or empty in config")
	}
	return parsed.Ansible.ImageKeys, parsed.Ansible.OtherImageKeys, nil
}
