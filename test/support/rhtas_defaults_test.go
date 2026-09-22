package support

import (
	"os"
	"reflect"
	"testing"
)

func TestAnsibleConfigOverrides(t *testing.T) {
	const disabled = `ansible:
  enabled: false
  imageKeys: [tas-image]
  otherImageKeys: [other-image]
`
	for _, testCase := range []struct {
		name     string
		defaults string
		override string
		enabled  bool
		wantErr  bool
	}{
		{name: "default disabled", defaults: disabled},
		{name: "legacy omitted flag", defaults: "ansible: {}", enabled: true},
		{name: "disabled without image lists", defaults: "ansible: {enabled: false}"},
		{name: "reenable future release", defaults: disabled, override: "rhtas: {ansible: {enabled: true}}", enabled: true},
		{name: "disable legacy config", defaults: "ansible: {}", override: "ansible: {enabled: false}"},
		{name: "partial override retains disabled", defaults: disabled, override: "rhtas: {ansible: {imageKeys: [tas-image]}}"},
		{name: "unwrapped enable", defaults: disabled, override: "ansible: {enabled: true}", enabled: true},
		{name: "invalid flag", defaults: disabled, override: "ansible: {enabled: invalid}", wantErr: true},
		{name: "invalid YAML", defaults: disabled, override: "ansible: [", wantErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			merged, err := MergeDefaultsConfig([]byte(testCase.defaults), []byte(testCase.override))
			var enabled bool
			if err == nil {
				enabled, err = GetAnsibleEnabledFromConfig(merged)
			}
			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected invalid Ansible configuration to fail")
				}
				return
			}
			if err != nil || enabled != testCase.enabled {
				t.Fatalf("enabled=%v, want %v, error=%v", enabled, testCase.enabled, err)
			}
			if testCase.defaults == disabled {
				images, otherImages, err := GetAnsibleImageKeysFromConfig(merged)
				if err != nil || !reflect.DeepEqual(images, []string{"tas-image"}) || !reflect.DeepEqual(otherImages, []string{"other-image"}) {
					t.Fatalf("partial override lost image lists: %v, %v, error=%v", images, otherImages, err)
				}
			}
		})
	}
}

func TestAnsibleReleaseConfigs(t *testing.T) {
	defaults, err := os.ReadFile("../acceptance/rhtas/defaults.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"1.5", "1.4", "1.3.2", "1.2.2"} {
		t.Run(version, func(t *testing.T) {
			var override []byte
			if version != "1.5" {
				var err error
				override, err = os.ReadFile("../../testdata/testconfig-" + version + ".yaml")
				if err != nil {
					t.Fatal(err)
				}
			}
			merged, err := MergeDefaultsConfig(defaults, override)
			if err != nil {
				t.Fatal(err)
			}
			enabled, err := GetAnsibleEnabledFromConfig(merged)
			if err != nil || enabled != (version != "1.5") {
				t.Fatalf("unexpected Ansible enablement: %v, error=%v", enabled, err)
			}
		})
	}
}
