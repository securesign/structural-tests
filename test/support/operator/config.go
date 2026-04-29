package operator

import (
	"errors"
	"fmt"

	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/config"
	"gopkg.in/yaml.v3"
)

type OperatorConfig struct {
	OperatorImageKey string            `yaml:"operatorImageKey"`
	Entrypoint       []string          `yaml:"entrypoint,omitempty"`
	Entrypointcmd    string            `yaml:"entrypointcmd"`
	ParseFormat      string            `yaml:"parseFormat"`
	ImageKeys        []string          `yaml:"imageKeys"`
	OtherImageKeys   []string          `yaml:"otherImageKeys,omitempty"`
	ImageKeyMap      map[string]string `yaml:"imageKeyMap,omitempty"`
	BundleImageKey   string            `yaml:"bundleImageKey"`
	BundleCSVPath    string            `yaml:"bundleCsvPath"`
}

type operatorSuiteSection struct {
	OperatorConfig `yaml:",inline"`
	Override       map[string]*OperatorConfig `yaml:"override,omitempty"`
}

func decodeOperatorSection(in interface{}) (operatorSuiteSection, error) {
	var out operatorSuiteSection
	if in == nil {
		return out, errors.New("operator section is nil")
	}
	conv := support.EnsureStringKeys(in)
	bytes, err := yaml.Marshal(conv)
	if err != nil {
		return out, fmt.Errorf("marshal operator section: %w", err)
	}
	if err := yaml.Unmarshal(bytes, &out); err != nil {
		return out, fmt.Errorf("decode operator section: %w", err)
	}
	backfillOperatorFromMap(conv, &out)
	return out, nil
}

func backfillOperatorFromMap(conv interface{}, out *operatorSuiteSection) {
	convMap, isMap := conv.(map[string]interface{})
	if !isMap {
		return
	}
	if out.OperatorImageKey == "" {
		if v, ok := convMap["operatorImageKey"].(string); ok {
			out.OperatorImageKey = v
		}
	}
	if out.Entrypointcmd == "" {
		if v, ok := convMap["entrypointcmd"].(string); ok {
			out.Entrypointcmd = v
		}
	}
	if out.ParseFormat == "" {
		if v, ok := convMap["parseFormat"].(string); ok {
			out.ParseFormat = v
		}
	}
	if out.BundleImageKey == "" {
		if v, ok := convMap["bundleImageKey"].(string); ok {
			out.BundleImageKey = v
		}
	}
	if out.BundleCSVPath == "" {
		if v, ok := convMap["bundleCsvPath"].(string); ok {
			out.BundleCSVPath = v
		}
	}
	backfillStringSlice(convMap, "imageKeys", &out.ImageKeys)
	backfillStringSlice(convMap, "otherImageKeys", &out.OtherImageKeys)
	backfillStringSlice(convMap, "entrypoint", &out.Entrypoint)
}

func backfillStringSlice(m map[string]interface{}, key string, target *[]string) {
	if len(*target) != 0 {
		return
	}
	raw, ok := m[key]
	if !ok {
		return
	}
	sl, isSlice := raw.([]interface{})
	if !isSlice {
		return
	}
	for _, item := range sl {
		if s, isStr := item.(string); isStr {
			*target = append(*target, s)
		}
	}
}

func getDefaultsOperator(defaultsData []byte) (operatorSuiteSection, error) {
	suiteMap, err := support.SuiteLevelMap(defaultsData)
	if err != nil {
		return operatorSuiteSection{}, fmt.Errorf("defaults for operator: %w", err)
	}
	operatorVal, ok := suiteMap["operator"]
	if !ok || operatorVal == nil {
		return operatorSuiteSection{}, errors.New("missing operator section in defaults")
	}
	return decodeOperatorSection(operatorVal)
}

func GetOperatorConfig(product string, defaultsData []byte) (OperatorConfig, error) {
	return getOperatorConfigBase(product, defaultsData)
}

func GetOperatorConfigForVersion(product, versionKey string, defaultsData []byte) (OperatorConfig, error) {
	base, err := getOperatorConfigBase(product, defaultsData)
	if err != nil {
		return OperatorConfig{}, err
	}
	defaultsOperator, err := getDefaultsOperator(defaultsData)
	if err != nil {
		return OperatorConfig{}, err
	}
	if defaultsOperator.Override != nil {
		if ov := defaultsOperator.Override[versionKey]; ov != nil {
			applyOperatorOverride(&base, ov)
		}
	}
	cfg, err := config.GetTestConfig()
	if err != nil {
		return OperatorConfig{}, fmt.Errorf("load test config: %w", err)
	}
	var userOperator operatorSuiteSection
	found, err := config.DecodeSection(cfg, product, "operator", &userOperator)
	if err != nil {
		return OperatorConfig{}, fmt.Errorf("decode operator section for %q: %w", product, err)
	}
	if found && userOperator.Override != nil {
		if ov := userOperator.Override[versionKey]; ov != nil {
			applyOperatorOverride(&base, ov)
		}
	}
	return base, nil
}

func getOperatorConfigBase(product string, defaultsData []byte) (OperatorConfig, error) {
	defaultsOperator, err := getDefaultsOperator(defaultsData)
	if err != nil {
		return OperatorConfig{}, fmt.Errorf("embedded operator defaults for %q: %w", product, err)
	}

	cfg, err := config.GetTestConfig()
	if err != nil {
		return OperatorConfig{}, fmt.Errorf("load test config: %w", err)
	}

	var userOperator operatorSuiteSection
	found, err := config.DecodeSection(cfg, product, "operator", &userOperator)
	if err != nil {
		return OperatorConfig{}, fmt.Errorf("decode operator section for %q: %w", product, err)
	}
	if found {
		applyOperatorDefaults(&userOperator.OperatorConfig, &defaultsOperator.OperatorConfig)
		return userOperator.OperatorConfig, nil
	}
	return defaultsOperator.OperatorConfig, nil
}

func applyOperatorDefaults(target, defaults *OperatorConfig) {
	if target.OperatorImageKey == "" {
		target.OperatorImageKey = defaults.OperatorImageKey
	}
	if target.Entrypointcmd == "" {
		target.Entrypointcmd = defaults.Entrypointcmd
	}
	if target.ParseFormat == "" {
		target.ParseFormat = defaults.ParseFormat
	}
	if target.BundleImageKey == "" {
		target.BundleImageKey = defaults.BundleImageKey
	}
	if target.BundleCSVPath == "" {
		target.BundleCSVPath = defaults.BundleCSVPath
	}
	if target.Entrypoint == nil {
		target.Entrypoint = defaults.Entrypoint
	}
	if target.ImageKeys == nil {
		target.ImageKeys = defaults.ImageKeys
	}
	if target.OtherImageKeys == nil {
		target.OtherImageKeys = defaults.OtherImageKeys
	}
	if target.ImageKeyMap == nil {
		target.ImageKeyMap = defaults.ImageKeyMap
	}
}

func applyOperatorOverride(base *OperatorConfig, override *OperatorConfig) {
	if override.OperatorImageKey != "" {
		base.OperatorImageKey = override.OperatorImageKey
	}
	if override.Entrypointcmd != "" {
		base.Entrypointcmd = override.Entrypointcmd
	}
	if override.ParseFormat != "" {
		base.ParseFormat = override.ParseFormat
	}
	if override.BundleImageKey != "" {
		base.BundleImageKey = override.BundleImageKey
	}
	if override.BundleCSVPath != "" {
		base.BundleCSVPath = override.BundleCSVPath
	}
	if override.Entrypoint != nil {
		base.Entrypoint = override.Entrypoint
	}
	if len(override.ImageKeys) > 0 {
		base.ImageKeys = override.ImageKeys
	}
	if override.OtherImageKeys != nil {
		base.OtherImageKeys = override.OtherImageKeys
	}
	if override.ImageKeyMap != nil {
		base.ImageKeyMap = override.ImageKeyMap
	}
}

func (c *OperatorConfig) SnapshotKey(imageKey string) string {
	if c.ImageKeyMap != nil {
		if mapped, ok := c.ImageKeyMap[imageKey]; ok {
			return mapped
		}
	}
	return imageKey
}
