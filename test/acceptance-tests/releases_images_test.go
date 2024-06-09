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

	It("snapshot.json file image snapshots are all unique", func() {
		snapshotHashes := support.ExtractHashes(support.GetMapValues(snapshotImages))
		mapped := make(map[string]int)
		for _, hash := range snapshotHashes {
			if _, ok := mapped[hash]; ok {
				mapped[hash]++
			} else {
				mapped[hash] = 1
			}
		}
		Expect(mapped).To(HaveEach(1))
		Expect(len(snapshotImages) == len(mapped)).To(BeTrue())
	})
})
