package fbc

import (
	"errors"
	"fmt"

	"github.com/securesign/structural-tests/test/support"
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

// fbcSuiteSection is the fbc suite: base fields plus override map (fbc.override in YAML).
type fbcSuiteSection struct {
	FBCConfig
	Override map[string]*FBCConfig `yaml:"override,omitempty"`
}

// ensureStringKeys converts map[interface{}]interface{} to map[string]interface{} recursively
// so yaml.Marshal produces correct keys (e.g. catalogPath) when decoding from parsed YAML.
func ensureStringKeys(input interface{}) interface{} {
	if input == nil {
		return nil
	}
	if m, ok := input.(map[interface{}]interface{}); ok {
		out := make(map[string]interface{}, len(m))
		for k, val := range m {
			if ks, ok := k.(string); ok {
				out[ks] = ensureStringKeys(val)
			}
		}
		return out
	}
	if s, ok := input.([]interface{}); ok {
		out := make([]interface{}, len(s))
		for i, val := range s {
			out[i] = ensureStringKeys(val)
		}
		return out
	}
	return input
}

func decodeFBCSection(in interface{}) (fbcSuiteSection, error) {
	var out fbcSuiteSection
	if in == nil {
		return out, errors.New("fbc section is nil")
	}
	conv := ensureStringKeys(in)
	bytes, err := yaml.Marshal(conv)
	if err != nil {
		return out, fmt.Errorf("marshal fbc section: %w", err)
	}
	if err := yaml.Unmarshal(bytes, &out); err != nil {
		return out, fmt.Errorf("decode fbc section: %w", err)
	}
	extractExpectedChannelsFromMap(conv, &out)
	return out, nil
}

// extractExpectedChannelsFromMap backfills ExpectedChannels from the map when unmarshal left it empty.
func extractExpectedChannelsFromMap(conv interface{}, out *fbcSuiteSection) {
	if len(out.ExpectedChannels) != 0 {
		return
	}
	convMap, isMap := conv.(map[string]interface{})
	if !isMap {
		return
	}
	ch, hasCh := convMap["expectedChannels"]
	if !hasCh {
		return
	}
	sl, isSlice := ch.([]interface{})
	if !isSlice {
		return
	}
	for _, item := range sl {
		if s, isStr := item.(string); isStr {
			out.ExpectedChannels = append(out.ExpectedChannels, s)
		}
	}
}

func getDefaultsFBC(defaultsData []byte) (fbcSuiteSection, error) {
	suiteMap, err := support.SuiteLevelMap(defaultsData)
	if err != nil {
		return fbcSuiteSection{}, fmt.Errorf("defaults for FBC: %w", err)
	}
	fbcVal, ok := suiteMap["fbc"]
	if !ok || fbcVal == nil {
		return fbcSuiteSection{}, errors.New("missing fbc section in defaults")
	}
	return decodeFBCSection(fbcVal)
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
// Uses shared fbc config, then applies fbc.override[versionKey] from defaults and TEST_CONFIG.
func GetFBCConfigForVersion(product, versionKey string, defaultsData []byte) (FBCConfig, error) {
	base, err := getFBCConfigBase(product, defaultsData)
	if err != nil {
		return FBCConfig{}, err
	}
	defaultsFBC, err := getDefaultsFBC(defaultsData)
	if err != nil {
		return FBCConfig{}, err
	}
	if defaultsFBC.Override != nil {
		if ov := defaultsFBC.Override[versionKey]; ov != nil {
			applyFBCOverride(&base, ov)
		}
	}
	cfg, err := config.GetTestConfig()
	if err != nil {
		return FBCConfig{}, fmt.Errorf("load test config: %w", err)
	}
	var userFBC fbcSuiteSection
	found, err := config.DecodeSection(cfg, product, "fbc", &userFBC)
	if err != nil {
		return FBCConfig{}, fmt.Errorf("decode fbc section for %q: %w", product, err)
	}
	if found && userFBC.Override != nil {
		if ov := userFBC.Override[versionKey]; ov != nil {
			applyFBCOverride(&base, ov)
		}
	}
	ensureFBCDefaults(product, &base)
	return base, nil
}

func getFBCConfigBase(product string, defaultsData []byte) (FBCConfig, error) {
	defaultsFBC, err := getDefaultsFBC(defaultsData)
	if err != nil {
		return FBCConfig{}, fmt.Errorf("embedded FBC defaults for %q: %w", product, err)
	}

	cfg, err := config.GetTestConfig()
	if err != nil {
		return FBCConfig{}, fmt.Errorf("load test config: %w", err)
	}

	var userFBC fbcSuiteSection
	found, err := config.DecodeSection(cfg, product, "fbc", &userFBC)
	if err != nil {
		return FBCConfig{}, fmt.Errorf("decode fbc section for %q: %w", product, err)
	}
	if found {
		applyFBCDefaults(&userFBC.FBCConfig, &defaultsFBC.FBCConfig)
		base := userFBC.FBCConfig
		ensureFBCDefaults(product, &base)
		return base, nil
	}
	base := defaultsFBC.FBCConfig
	ensureFBCDefaults(product, &base)
	return base, nil
}

// ensureFBCDefaults sets required defaults so FBC tests never use wrong images or empty paths.
// Without ImageKeyPrefix we would match all snapshot images (e.g. backfill-redis) as "FBC" and fail.
func ensureFBCDefaults(product string, base *FBCConfig) {
	if product != "rhtas" {
		return
	}
	if base.CatalogPath == "" {
		base.CatalogPath = "/configs/rhtas-operator/catalog.json"
	}
	if base.ImageKeyPrefix == "" {
		base.ImageKeyPrefix = "fbc-"
	}
	if base.OLMPackage == "" {
		base.OLMPackage = "rhtas-operator"
	}
	if base.OperatorBundleImage == "" {
		base.OperatorBundleImage = "registry.redhat.io/rhtas/rhtas-operator-bundle"
	}
	if base.DefaultChannel == "" {
		base.DefaultChannel = "stable"
	}
	if len(base.ExpectedChannels) == 0 {
		// Default set matches FBC catalog (stable + versioned channels), e.g. 1.3.x and 1.4.0.
		base.ExpectedChannels = []string{"stable", "stable-v1.1", "stable-v1.2", "stable-v1.3", "stable-v1.4"}
	}
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
