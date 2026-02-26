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
	FBC          *FBCConfig            `yaml:"fbc"`
	FBCOverrides map[string]*FBCConfig `yaml:"fbcOverrides,omitempty"`
}

// GetFBCConfig returns FBC config for the given product (shared config).
// Embedded defaultsData provides fallback values; TEST_CONFIG overrides them.
func GetFBCConfig(product string, defaultsData []byte) (FBCConfig, error) {
	base, err := getFBCConfigBase(product, defaultsData)
	if err != nil {
		return FBCConfig{}, err
	}
	return base, nil
}

// GetFBCConfigForVersion returns FBC config for the given product and version key.
// Uses shared fbc config, then applies fbcOverrides[versionKey] from defaults.yaml if present.
func GetFBCConfigForVersion(product, versionKey string, defaultsData []byte) (FBCConfig, error) {
	base, err := getFBCConfigBase(product, defaultsData)
	if err != nil {
		return FBCConfig{}, err
	}
	var defaults productDefaults
	if err := yaml.Unmarshal(defaultsData, &defaults); err != nil {
		return FBCConfig{}, fmt.Errorf("parse embedded FBC defaults for %q: %w", product, err)
	}
	if defaults.FBCOverrides != nil {
		if override := defaults.FBCOverrides[versionKey]; override != nil {
			applyFBCOverride(&base, override)
		}
	}
	return base, nil
}

func getFBCConfigBase(product string, defaultsData []byte) (FBCConfig, error) {
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

// applyFBCOverride applies version-specific overrides onto base (override wins for set fields).
func applyFBCOverride(base *FBCConfig, override *FBCConfig) {
	if override.OLMPackage != "" {
		base.OLMPackage = override.OLMPackage
	}
	if override.OperatorBundleImage != "" {
		base.OperatorBundleImage = override.OperatorBundleImage
	}
	if override.CatalogPath != "" {
		base.CatalogPath = override.CatalogPath
	}
	if override.ImageKeyPrefix != "" {
		base.ImageKeyPrefix = override.ImageKeyPrefix
	}
	if override.DefaultChannel != "" {
		base.DefaultChannel = override.DefaultChannel
	}
	if len(override.ExpectedChannels) > 0 {
		base.ExpectedChannels = override.ExpectedChannels
	}
	if override.ExpectedDeprecations != nil {
		base.ExpectedDeprecations = override.ExpectedDeprecations
	}
}
