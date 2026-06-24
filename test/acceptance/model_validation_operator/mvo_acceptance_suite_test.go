package acceptance

import (
	_ "embed"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const product = "model_validation_operator"

//go:embed defaults.yaml
var defaults []byte

func TestAcceptance(t *testing.T) {
	format.MaxLength = 0

	RegisterFailHandler(Fail)
	log.SetLogger(GinkgoLogr)
	RunSpecs(t, "Model Validation Operator Acceptance Tests Suite")
}
