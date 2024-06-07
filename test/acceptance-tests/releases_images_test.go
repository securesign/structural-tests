package acceptance_tests

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
)

var _ = Describe("Trusted Artifact Signer Releases", Ordered, func() {

	var (
		snapshotImages support.SnapshotMap
	)

	It("snapshot.json file exist and is parseable", func() {
		content, err := support.GetFileContent(support.GetEnvOrDefault(support.EnvReleasesSnapshotFile, support.DefaultReleasesSnapshotFile))
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(BeEmpty())

		err = json.Unmarshal([]byte(content), &snapshotImages)
		Expect(err).NotTo(HaveOccurred())
		log.Printf("Found %d snapshot images\n", len(snapshotImages))
	})

	It("snapshot.json file contains valid images", func() {
		Expect(snapshotImages).NotTo(BeEmpty())
		Expect(snapshotImages).NotTo(ContainElement(BeEmpty()))
		Expect(snapshotImages).To(HaveEach(MatchRegexp(support.ImageDefinitionRegexp)))
	})

	It("snapshot.json file images are all unique", func() {
		existingImages := make(map[string]struct{})
		for _, image := range support.GetMapValues(snapshotImages) {
			if _, ok := existingImages[image]; ok {
				Fail("Not unique image: " + image)
			}
			existingImages[image] = struct{}{}
		}
	})
})
