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
	FBCConfig `yaml:",inline"`
	Override  map[string]*FBCConfig `yaml:"override,omitempty"`
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
	backfillFBCFromMap(conv, &out)
	return out, nil
}

// backfillFBCFromMap sets FBC struct fields from the config map when unmarshal left them empty.
// This fixes decoding when the source is a nested map (e.g. embedded defaults) and avoids hardcoding.
func backfillFBCFromMap(conv interface{}, out *fbcSuiteSection) {
	convMap, isMap := conv.(map[string]interface{})
	if !isMap {
		return
	}
	if out.CatalogPath == "" {
		if v, ok := convMap["catalogPath"].(string); ok {
			out.CatalogPath = v
		}
	}
	if out.ImageKeyPrefix == "" {
		if v, ok := convMap["imageKeyPrefix"].(string); ok {
			out.ImageKeyPrefix = v
		}
	}
	if out.OLMPackage == "" {
		if v, ok := convMap["olmPackage"].(string); ok {
			out.OLMPackage = v
		}
	}
	if out.OperatorBundleImage == "" {
		if v, ok := convMap["operatorBundleImage"].(string); ok {
			out.OperatorBundleImage = v
		}
	}
	if out.DefaultChannel == "" {
		if v, ok := convMap["defaultChannel"].(string); ok {
			out.DefaultChannel = v
		}
	}
	extractExpectedChannelsFromMap(conv, out)
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
		return base, nil
	}
	base := defaultsFBC.FBCConfig
	return base, nil
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
