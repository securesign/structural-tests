package acceptance

import (
	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/operator"
)

func operatorDefaults() []byte {
	content, err := support.GetTestConfigContent()
	if err != nil || len(content) == 0 {
		return defaults
	}
	merged, err := support.MergeDefaultsConfig(defaults, content)
	if err != nil {
		return defaults
	}
	return merged
}

var _ = operator.DescribeOperatorImageTests(product, operatorDefaults())
