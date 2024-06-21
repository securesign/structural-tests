package acceptance

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestAcceptance(t *testing.T) {
	format.MaxLength = 0

	RegisterFailHandler(Fail)
	log.SetLogger(GinkgoLogr)
	RunSpecs(t, "Trusted Artifact Signer Acceptance Tests Suite")
}
