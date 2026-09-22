package cli

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/config"
	"gopkg.in/yaml.v3"
)

type StackConfig struct {
	Images []StackImage `yaml:"images"`
}

type StackImage struct {
	ImageKey string        `yaml:"imageKey"`
	Binaries []StackBinary `yaml:"binaries"`
}

type StackBinary struct {
	Path string `yaml:"path"`
	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

// GetCLIStackConfig overlays the product's TEST_CONFIG inventory on embedded defaults.
// An omitted images field inherits defaults; an explicit empty list disables checks.
func GetCLIStackConfig(product string, defaultsData []byte) (StackConfig, error) {
	defaults, err := support.SuiteLevelMap(defaultsData)
	if err != nil {
		return StackConfig{}, fmt.Errorf("CLI stack defaults for %q: %w", product, err)
	}
	base, err := decodeStackConfig(defaults["cliStack"])
	if err != nil {
		return base, fmt.Errorf("CLI stack defaults for %q: %w", product, err)
	}
	user, err := config.GetTestConfig()
	if err != nil {
		return base, fmt.Errorf("load CLI stack config: %w", err)
	}
	if section, found := user[product]["cliStack"]; found {
		override, decodeErr := decodeStackConfig(section)
		if decodeErr != nil {
			return base, fmt.Errorf("CLI stack config for %q: %w", product, decodeErr)
		}
		if override.Images != nil {
			base.Images = override.Images
		}
	}
	return base, base.validate()
}

func decodeStackConfig(section interface{}) (StackConfig, error) {
	var result StackConfig
	if section == nil {
		return result, errors.New("cliStack section must be a mapping")
	}
	data, err := yaml.Marshal(section)
	if err != nil {
		return result, fmt.Errorf("marshal cliStack: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("decode cliStack: %w", err)
	}
	return result, nil
}

func (cfg StackConfig) validate() error {
	for _, image := range cfg.Images {
		if strings.TrimSpace(image.ImageKey) == "" || len(image.Binaries) == 0 {
			return errors.New("cliStack images require imageKey and nonempty binaries")
		}
		for _, binary := range image.Binaries {
			if !strings.HasPrefix(binary.Path, "/binaries/") || path.Clean(binary.Path) != binary.Path ||
				!strings.HasSuffix(binary.Path, ".gz") {
				return fmt.Errorf("CLI stack %q: invalid archive path %q", image.ImageKey, binary.Path)
			}
			if !validPlatform(binary.OS, binary.Arch) {
				return fmt.Errorf("CLI stack %q: unsupported platform %s/%s", image.ImageKey, binary.OS, binary.Arch)
			}
		}
	}
	return nil
}

func validPlatform(osName, arch string) bool {
	switch osName {
	case "linux":
		return arch == archAMD64 || arch == archARM64 || arch == "ppc64le" || arch == "s390x"
	case "darwin", "windows":
		return arch == archAMD64 || arch == archARM64
	default:
		return false
	}
}
