package acceptance_tests

import (
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
		var err error
		snapshotImages, err = support.ParseSnapshotImages()
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotImages).NotTo(BeEmpty())
		log.Printf("Found %d snapshot images\n", len(snapshotImages))
	})

	It("snapshot.json file contains valid images", func() {
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
