package acceptance

import (
	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/fbc"
)

func fbcDefaults() []byte {
	if content, err := support.GetTestConfigContent(); err == nil && len(content) > 0 {
		return content
	}
	return defaults
}

var _ = fbc.DescribeFBCImageTests(product, fbcDefaults())
