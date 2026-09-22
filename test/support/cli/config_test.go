package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/securesign/structural-tests/test/support"
)

const testDefaults = `cliStack:
  images:
    - imageKey: tufcli-cli-stack-image
      binaries:
        - {path: /binaries/tufcli_linux_amd64.tar.gz, os: linux, arch: amd64}
`

func TestGetCLIStackConfig(t *testing.T) {
	legacy := strings.ReplaceAll(testDefaults, "tufcli", "tuftool")
	wrapped := func(product, section string) string {
		return product + ":\n  " + strings.ReplaceAll(strings.TrimSpace(section), "\n", "\n  ") + "\n"
	}
	for _, testCase := range []struct {
		name    string
		product string
		yaml    string
		want    string
		wantErr bool
	}{
		{name: "defaults", product: "rhtas", want: "tufcli-cli-stack-image"},
		{name: "replacement", product: "rhtas", yaml: wrapped("rhtas", legacy), want: "tuftool-cli-stack-image"},
		{name: "unwrapped", product: "rhtas", yaml: legacy, want: "tuftool-cli-stack-image"},
		{name: "unwrapped with operator", product: "rhtas", yaml: "operator: {}\n" + legacy, want: "tuftool-cli-stack-image"},
		{name: "disabled", product: "rhtas", yaml: "rhtas: {cliStack: {images: []}}"},
		{name: "omitted list", product: "rhtas", yaml: "rhtas: {cliStack: {}}", want: "tufcli-cli-stack-image"},
		{name: "other product", product: "rhtas", yaml: wrapped("policy_controller", legacy), want: "tufcli-cli-stack-image"},
		{name: "policy inventory", product: "policy_controller", yaml: wrapped("policy_controller", legacy), want: "tuftool-cli-stack-image"},
		{name: "unwrapped is RHTAS only", product: "model_transparency", yaml: legacy, want: "tufcli-cli-stack-image"},
		{name: "bad YAML", product: "rhtas", yaml: "rhtas: [", wantErr: true},
		{name: "bad product type", product: "rhtas", yaml: "rhtas: false", wantErr: true},
		{name: "unknown field", product: "rhtas", yaml: "cliStack: {imagez: []}", wantErr: true},
		{name: "null section", product: "rhtas", yaml: "cliStack: null", wantErr: true},
		{name: "wrong list type", product: "rhtas", yaml: "cliStack: {images: false}", wantErr: true},
		{name: "missing key", product: "rhtas", yaml: strings.ReplaceAll(testDefaults, "imageKey:", "imageKey: #"), wantErr: true},
		{name: "empty binaries", product: "rhtas", yaml: "cliStack: {images: [{imageKey: cli-image, binaries: []}]}", wantErr: true},
		{name: "invalid OS", product: "rhtas", yaml: strings.ReplaceAll(testDefaults, "os: linux", "os: unknown"), wantErr: true},
		{name: "invalid arch", product: "rhtas", yaml: strings.ReplaceAll(testDefaults, "arch: amd64", "arch: unknown"), wantErr: true},
		{name: "unsupported pair", product: "rhtas",
			yaml: strings.ReplaceAll(strings.ReplaceAll(testDefaults, "os: linux", "os: darwin"), "arch: amd64", "arch: s390x"), wantErr: true},
		{name: "path traversal", product: "rhtas", yaml: strings.ReplaceAll(testDefaults, "/binaries/", "/binaries/../"), wantErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(support.EnvTestConfig, "")
			if testCase.yaml != "" {
				setTestConfig(t, testCase.yaml)
			}
			cfg, err := GetCLIStackConfig(testCase.product, []byte(testDefaults))
			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected invalid configuration to fail")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if testCase.want == "" {
				if len(cfg.Images) != 0 {
					t.Fatalf("disabled inventory has %d images", len(cfg.Images))
				}
				return
			}
			if len(cfg.Images) != 1 || cfg.Images[0].ImageKey != testCase.want {
				t.Fatalf("expected only %s, got %+v", testCase.want, cfg.Images)
			}
		})
	}
}

func TestReleaseInventories(t *testing.T) {
	for _, testCase := range []struct {
		product string
		name    string
		config  string
		keys    []string
		count   int
		parity  int
	}{
		{product: "rhtas", name: "1.5", keys: []string{
			"cosign-cli-stack-image", "gitsign-cli-stack-image", "rekor-cli-stack-image", "fetch-tsa-certs-cli-stack-image",
			"trillian-cli-stack-image", "tufcli-cli-stack-image", "conforma-cli-stack-image",
		}, count: 56, parity: 28},
		{product: "rhtas", name: "1.4", config: "testdata/testconfig-1.4.yaml", keys: []string{
			"cosign-cli-stack-image", "gitsign-cli-stack-image", "rekor-cli-stack-image", "fetch-tsa-certs-cli-stack-image",
			"trillian-cli-stack-image", "tuftool-cli-stack-image", "conforma-cli-stack-image", "model-transparency-cli-stack-image",
		}, count: 65, parity: 25},
		{product: "rhtas", name: "1.3", config: "testdata/testconfig-1.3.2.yaml"},
		{product: "rhtas", name: "1.2", config: "testdata/testconfig-1.2.2.yaml"},
		{product: "model_transparency", name: "defaults", keys: []string{"model-transparency-cli-stack-image"}, count: 15},
		{product: "policy_controller", name: "defaults"},
	} {
		t.Run(testCase.product+"/"+testCase.name, func(t *testing.T) {
			t.Setenv(support.EnvTestConfig, testCase.config)
			defaults, err := os.ReadFile("../../acceptance/" + testCase.product + "/defaults.yaml")
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := GetCLIStackConfig(testCase.product, defaults)
			if err != nil {
				t.Fatal(err)
			}
			var keys []string
			count, parity := 0, 0
			for _, image := range cfg.Images {
				keys = append(keys, image.ImageKey)
				count += len(image.Binaries)
				for _, binary := range image.Binaries {
					if binary.Runtime != nil {
						parity++
					}
				}
			}
			if !reflect.DeepEqual(keys, testCase.keys) || count != testCase.count || parity != testCase.parity {
				t.Fatalf("unexpected inventory: keys=%v, archives=%d, parity=%d", keys, count, parity)
			}
		})
	}
}

func setTestConfig(t *testing.T, content string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "testconfig.yaml")
	if err := os.WriteFile(file, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(support.EnvTestConfig, file)
}

func TestRuntimeConfiguration(t *testing.T) {
	withRuntime := strings.ReplaceAll(testDefaults, "arch: amd64}", "arch: amd64, runtime: {imageKey: tufcli-image, path: /tufcli}}")
	for _, testCase := range []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{"valid", withRuntime, false},
		{"archive only", testDefaults, false},
		{"empty mapping", strings.ReplaceAll(withRuntime, "imageKey: tufcli-image, path: /tufcli", ""), true},
		{"missing path", strings.ReplaceAll(withRuntime, ", path: /tufcli", ""), true},
		{"missing image", strings.ReplaceAll(withRuntime, "imageKey: tufcli-image, ", ""), true},
		{"relative path", strings.ReplaceAll(withRuntime, "path: /tufcli", "path: tufcli"), true},
		{"unclean path", strings.ReplaceAll(withRuntime, "path: /tufcli", "path: /bin/../tufcli"), true},
		{"root path", strings.ReplaceAll(withRuntime, "path: /tufcli", "path: /"), true},
		{"non Linux", strings.ReplaceAll(withRuntime, "os: linux", "os: darwin"), true},
		{"unknown field", strings.ReplaceAll(withRuntime, "path: /tufcli", "file: /tufcli"), true},
		{"wrong type", strings.ReplaceAll(withRuntime, "{imageKey: tufcli-image, path: /tufcli}", "false"), true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			setTestConfig(t, testCase.yaml)
			_, err := GetCLIStackConfig("rhtas", []byte(testDefaults))
			if (err != nil) != testCase.wantErr {
				t.Fatalf("error=%v, wantErr=%v", err, testCase.wantErr)
			}
		})
	}
	for _, testCase := range []struct {
		name, config, product string
		want                  bool
	}{
		{"inherited", "", "rhtas", true},
		{"removed by replacement", testDefaults, "rhtas", false},
		{"unwrapped mapping", withRuntime, "rhtas", true},
		{"other product", "policy_controller: {cliStack: {images: []}}", "rhtas", true},
		{"disabled", "cliStack: {images: []}", "rhtas", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(support.EnvTestConfig, "")
			if testCase.config != "" {
				setTestConfig(t, testCase.config)
			}
			cfg, err := GetCLIStackConfig(testCase.product, []byte(withRuntime))
			if err != nil {
				t.Fatal(err)
			}
			got := len(cfg.Images) > 0 && cfg.Images[0].Binaries[0].Runtime != nil
			if got != testCase.want {
				t.Fatalf("runtime present=%v, want %v", got, testCase.want)
			}
		})
	}
}
