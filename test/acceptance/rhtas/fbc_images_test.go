package acceptance

import (
	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/fbc"
)

func fbcDefaults() []byte {
	content, err := support.GetTestConfigContent()
	if err != nil || len(content) == 0 {
		return defaults
	}
	merged, err := support.MergeRhtasConfig(defaults, content)
	if err != nil {
		return defaults
	}
	return merged
}

var _ = fbc.DescribeFBCImageTests(product, fbcDefaults())
