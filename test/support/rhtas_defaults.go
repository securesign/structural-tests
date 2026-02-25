package support

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// rhtasDefaults is the subset of test/acceptance/rhtas/defaults.yaml used for operator and ansible image keys.
type rhtasDefaults struct {
	MandatoryTasOperatorImageKeys           []string                     `yaml:"mandatoryTasOperatorImageKeys"`
	MandatoryTasOperatorImageKeysOverrides  map[string][]string          `yaml:"mandatoryTasOperatorImageKeysOverrides,omitempty"`
	AnsibleTasImageKeys                     []string                     `yaml:"ansibleTasImageKeys"`
	AnsibleOtherImageKeys                   []string                     `yaml:"ansibleOtherImageKeys"`
	AnsibleImageKeysOverrides               map[string]ansibleKeysEntry `yaml:"ansibleImageKeysOverrides,omitempty"`
}

// ansibleKeysEntry is the per-version override for ansible TAS and other image keys.
type ansibleKeysEntry struct {
	AnsibleTasImageKeys   []string `yaml:"ansibleTasImageKeys"`
	AnsibleOtherImageKeys []string `yaml:"ansibleOtherImageKeys"`
}

// GetMandatoryTasOperatorImageKeysFromConfig returns the mandatory TAS operator image key list from
// rhtas defaults.yaml. version is from VERSION env (e.g. "1.2", "1.3.2"); overrides are keyed by
// exact version or major.minor (e.g. "1.2" matches "1.2.0"). If defaultsYaml is nil/empty or the
// key is missing, returns MandatoryTasOperatorImageKeys() as fallback.
func GetMandatoryTasOperatorImageKeysFromConfig(defaultsYaml []byte, version string) ([]string, error) {
	if len(defaultsYaml) == 0 {
		return MandatoryTasOperatorImageKeys(), nil
	}
	var d rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &d); err != nil {
		return nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(d.MandatoryTasOperatorImageKeys) == 0 {
		return MandatoryTasOperatorImageKeys(), nil
	}
	// Prefer version-specific override: exact match, then major.minor
	if version != "" && d.MandatoryTasOperatorImageKeysOverrides != nil {
		if keys, ok := d.MandatoryTasOperatorImageKeysOverrides[version]; ok && len(keys) > 0 {
			return keys, nil
		}
		if majorMinor := majorMinorVersion(version); majorMinor != "" && majorMinor != version {
			if keys, ok := d.MandatoryTasOperatorImageKeysOverrides[majorMinor]; ok && len(keys) > 0 {
				return keys, nil
			}
		}
	}
	return d.MandatoryTasOperatorImageKeys, nil
}

// GetAnsibleTasImageKeysFromConfig returns the ansible TAS image key list from rhtas defaults.yaml.
// version is from VERSION env. If defaultsYaml is nil/empty or the key is missing, returns AnsibleTasImageKeys() as fallback.
func GetAnsibleTasImageKeysFromConfig(defaultsYaml []byte, version string) ([]string, error) {
	tas, _, err := GetAnsibleImageKeysFromConfig(defaultsYaml, version)
	return tas, err
}

// GetAnsibleOtherImageKeysFromConfig returns the ansible other image key list from rhtas defaults.yaml.
// version is from VERSION env. If defaultsYaml is nil/empty or the key is missing, returns AnsibleOtherImageKeys() as fallback.
func GetAnsibleOtherImageKeysFromConfig(defaultsYaml []byte, version string) ([]string, error) {
	_, other, err := GetAnsibleImageKeysFromConfig(defaultsYaml, version)
	return other, err
}

// GetAnsibleImageKeysFromConfig returns both ansible TAS and other image key lists from rhtas defaults.yaml.
func GetAnsibleImageKeysFromConfig(defaultsYaml []byte, version string) (tasKeys, otherKeys []string, err error) {
	return getAnsibleImageKeysFromConfig(defaultsYaml, version)
}

func getAnsibleImageKeysFromConfig(defaultsYaml []byte, version string) (tasKeys, otherKeys []string, err error) {
	if len(defaultsYaml) == 0 {
		return AnsibleTasImageKeys(), AnsibleOtherImageKeys(), nil
	}
	var d rhtasDefaults
	if err := yaml.Unmarshal(defaultsYaml, &d); err != nil {
		return nil, nil, fmt.Errorf("parse rhtas defaults: %w", err)
	}
	if len(d.AnsibleTasImageKeys) == 0 {
		return AnsibleTasImageKeys(), AnsibleOtherImageKeys(), nil
	}
	if len(d.AnsibleOtherImageKeys) == 0 {
		return d.AnsibleTasImageKeys, AnsibleOtherImageKeys(), nil
	}
	// Prefer version-specific override
	if version != "" && d.AnsibleImageKeysOverrides != nil {
		if entry, ok := d.AnsibleImageKeysOverrides[version]; ok && (len(entry.AnsibleTasImageKeys) > 0 || len(entry.AnsibleOtherImageKeys) > 0) {
			tas := entry.AnsibleTasImageKeys
			other := entry.AnsibleOtherImageKeys
			if len(tas) == 0 {
				tas = d.AnsibleTasImageKeys
			}
			if len(other) == 0 {
				other = d.AnsibleOtherImageKeys
			}
			return tas, other, nil
		}
		if majorMinor := majorMinorVersion(version); majorMinor != "" && majorMinor != version {
			if entry, ok := d.AnsibleImageKeysOverrides[majorMinor]; ok && (len(entry.AnsibleTasImageKeys) > 0 || len(entry.AnsibleOtherImageKeys) > 0) {
				tas := entry.AnsibleTasImageKeys
				other := entry.AnsibleOtherImageKeys
				if len(tas) == 0 {
					tas = d.AnsibleTasImageKeys
				}
				if len(other) == 0 {
					other = d.AnsibleOtherImageKeys
				}
				return tas, other, nil
			}
		}
	}
	return d.AnsibleTasImageKeys, d.AnsibleOtherImageKeys, nil
}

func majorMinorVersion(v string) string {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "." + parts[1]
}
