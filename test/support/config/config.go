package config

import (
	"fmt"

	"github.com/securesign/structural-tests/test/support"
	"gopkg.in/yaml.v3"
)

// TestConfig maps product names to their configuration sections.
type TestConfig map[string]map[string]interface{}

// GetTestConfig returns test configuration from TEST_CONFIG env or the default path.
func GetTestConfig() (TestConfig, error) {
	return resolveTestConfig()
}

func resolveTestConfig() (TestConfig, error) {
	path := support.GetEnv(support.EnvTestConfig)
	if path == "" {
		path = support.DefaultTestConfigPath
	}

	content, err := support.GetFileContent(path)
	if err != nil {
		return nil, fmt.Errorf("read test config %s: %w", path, err)
	}

	var cfg TestConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse test config %s: %w", path, err)
	}
	return cfg, nil
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
	raw, err := yaml.Marshal(sectionData)
	if err != nil {
		return false, fmt.Errorf("re-marshal section %q of product %q: %w", section, product, err)
	}
	if err := yaml.Unmarshal(raw, target); err != nil {
		return false, fmt.Errorf("decode section %q of product %q: %w", section, product, err)
	}
	return true, nil
}
