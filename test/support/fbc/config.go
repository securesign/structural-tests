package fbc

import (
	"fmt"

	"github.com/securesign/structural-tests/test/support/config"
	"gopkg.in/yaml.v3"
)

// FBCConfig holds file-based catalog test parameters.
type FBCConfig struct {
	OLMPackage           string   `yaml:"olmPackage"`
	OperatorBundleImage  string   `yaml:"operatorBundleImage"`
	CatalogPath          string   `yaml:"catalogPath"`
	ImageKeyPrefix       string   `yaml:"imageKeyPrefix"`
	DefaultChannel       string   `yaml:"defaultChannel"`
	ExpectedChannels     []string `yaml:"expectedChannels"`
	ExpectedDeprecations []string `yaml:"expectedDeprecations,omitempty"`
}

type productDefaults struct {
	FBC *FBCConfig `yaml:"fbc"`
}

// GetFBCConfig returns FBC config for the given product.
// Embedded defaultsData provides fallback values; TEST_CONFIG overrides them.
func GetFBCConfig(product string, defaultsData []byte) (FBCConfig, error) {
	var defaults productDefaults
	if err := yaml.Unmarshal(defaultsData, &defaults); err != nil {
		return FBCConfig{}, fmt.Errorf("parse embedded FBC defaults for %q: %w", product, err)
	}
	if defaults.FBC == nil {
		return FBCConfig{}, fmt.Errorf("missing fbc section in embedded defaults for %q", product)
	}

	cfg, err := config.GetTestConfig()
	if err != nil {
		return FBCConfig{}, fmt.Errorf("load test config: %w", err)
	}

	var userFBC FBCConfig
	found, err := config.DecodeSection(cfg, product, "fbc", &userFBC)
	if err != nil {
		return FBCConfig{}, fmt.Errorf("decode fbc section for %q: %w", product, err)
	}
	if found {
		applyFBCDefaults(&userFBC, defaults.FBC)
		return userFBC, nil
	}
	return *defaults.FBC, nil
}

func applyFBCDefaults(from, defaults *FBCConfig) {
	if from.OLMPackage == "" {
		from.OLMPackage = defaults.OLMPackage
	}
	if from.OperatorBundleImage == "" {
		from.OperatorBundleImage = defaults.OperatorBundleImage
	}
	if from.CatalogPath == "" {
		from.CatalogPath = defaults.CatalogPath
	}
	if from.ImageKeyPrefix == "" {
		from.ImageKeyPrefix = defaults.ImageKeyPrefix
	}
	if from.DefaultChannel == "" {
		from.DefaultChannel = defaults.DefaultChannel
	}
	if from.ExpectedChannels == nil {
		from.ExpectedChannels = defaults.ExpectedChannels
	}
	if from.ExpectedDeprecations == nil {
		from.ExpectedDeprecations = defaults.ExpectedDeprecations
	}
}
