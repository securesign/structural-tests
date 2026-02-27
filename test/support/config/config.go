package config

import (
	"fmt"

	"github.com/securesign/structural-tests/test/support"
	"gopkg.in/yaml.v3"
)

// TestConfig maps product names to their configuration sections.
type TestConfig map[string]map[string]interface{}

// GetTestConfig returns test configuration from TEST_CONFIG env or the default path.
// Accepts both formats: (1) product -> section -> data (e.g. rhtas.fbc), and (2) flat defaults
// with top-level fbc, fbcOverrides, and list keys (mandatoryTasOperatorImageKeys, etc.).
// In format (2), top-level "fbc" is treated as rhtas.fbc so FBC config is still found.
func GetTestConfig() (TestConfig, error) {
	return resolveTestConfig()
}

func resolveTestConfig() (TestConfig, error) {
	path := support.GetEnv(support.EnvTestConfig)
	if path == "" {
		return TestConfig{}, nil
	}

	content, err := support.GetFileContent(path)
	if err != nil {
		return nil, fmt.Errorf("read test config %s: %w", path, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("parse test config %s: %w", path, err)
	}

	cfg := make(TestConfig)
	if raw == nil {
		return cfg, nil
	}

	if rhtas, ok := raw["rhtas"]; ok {
		if rhtasMap, ok := toMap(rhtas); ok {
			cfg["rhtas"] = rhtasMap
			return cfg, nil
		}
	}
	// Suite-level file (operator, ansible, fbc at top level) without package wrapper
	if _, hasOperator := raw["operator"]; hasOperator {
		cfg["rhtas"] = raw
		return cfg, nil
	}
	if _, hasAnsible := raw["ansible"]; hasAnsible {
		cfg["rhtas"] = raw
		return cfg, nil
	}
	if fbc, ok := raw["fbc"]; ok {
		if fbcMap, ok := toMap(fbc); ok {
			cfg["rhtas"] = map[string]interface{}{"fbc": fbcMap}
			return cfg, nil
		}
	}
	return cfg, nil
}

func toMap(value interface{}) (map[string]interface{}, bool) {
	if value == nil {
		return nil, false
	}
	m, ok := value.(map[string]interface{})
	if ok {
		return m, true
	}
	// yaml.Unmarshal may produce map[interface{}]interface{}
	if m2, ok := value.(map[interface{}]interface{}); ok {
		out := make(map[string]interface{}, len(m2))
		for k, val := range m2 {
			if ks, ok := k.(string); ok {
				out[ks] = val
			}
		}
		return out, true
	}
	return nil, false
}

// toDeepStringMap recursively converts map[interface{}]interface{} to map[string]interface{}
// so YAML marshal produces correct keys (e.g. catalogPath).
func toDeepStringMap(input interface{}) interface{} {
	if input == nil {
		return nil
	}
	if m, ok := input.(map[interface{}]interface{}); ok {
		out := make(map[string]interface{}, len(m))
		for k, val := range m {
			if ks, ok := k.(string); ok {
				out[ks] = toDeepStringMap(val)
			}
		}
		return out
	}
	if s, ok := input.([]interface{}); ok {
		out := make([]interface{}, len(s))
		for i, val := range s {
			out[i] = toDeepStringMap(val)
		}
		return out
	}
	return input
}

// DecodeSection unmarshals a product's configuration section into target.
// Returns false if the product or section is not present.
func DecodeSection(cfg TestConfig, product, section string, target interface{}) (bool, error) {
	prodMap, ok := cfg[product]
	if !ok {
		return false, nil
	}
	sectionData, ok := prodMap[section]
	if !ok {
		return false, nil
	}
	sectionData = toDeepStringMap(sectionData)
	raw, err := yaml.Marshal(sectionData)
	if err != nil {
		return false, fmt.Errorf("re-marshal section %q of product %q: %w", section, product, err)
	}
	if err := yaml.Unmarshal(raw, target); err != nil {
		return false, fmt.Errorf("decode section %q of product %q: %w", section, product, err)
	}
	return true, nil
}
