package acceptance_tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
	"strings"
)

var _ = Describe("Trusted Artifact Signer Operator", Ordered, func() {

	var (
		err               error
		imagesFileContent string
		operatorImages    []string
	)

	It("read images.go file from operator repository", func() {
		operatorBranch := support.GetEnvOrDefault(support.EnvOperatorRepoBranch, support.OperatorRepoDefBranch)
		imagesFileContent, err = support.DownloadFileContent(fmt.Sprintf(support.OperatorImagesFile, operatorBranch), "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("extract image definitions", func() {
		operatorImages, err = support.GetImageDefinitionsFromConstants(imagesFileContent)
		Expect(err).NotTo(HaveOccurred())
		log.Printf("Found %d images: \n%v", len(operatorImages), strings.Join(operatorImages, "\n"))
	})

	It("image definitions are pointing to Red Hat registry", func() {
		for _, image := range operatorImages {
			Expect(image).To(Or(
				HavePrefix("registry.redhat.io/"),
				HavePrefix("registry.access.redhat.com/"),
			))
		}
	})

	It("image definitions are all unique", func() {
		existingImages := make(map[string]struct{})
		for _, image := range operatorImages {
			if _, ok := existingImages[image]; ok {
				Fail("Not unique image: " + image)
			}
			existingImages[image] = struct{}{}
		}
	})

	It("image definitions have hashes as tags", func() {
		for _, image := range operatorImages {
			parts := strings.Split(image, "@sha256:")
			Expect(parts).To(HaveLen(2), fmt.Sprintf("Not a hash tag for %s", image))
			Expect(support.HasValidSHA(parts[1])).To(BeTrue(), fmt.Sprintf("Image has invalid hash for %s", image))
		}
	})
})
