package acceptance

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

var _ = Describe("Policy Controller Operator Releases", Ordered, func() {

	var (
		snapshotData support.SnapshotData
	)

	It("snapshot.json file exist and is parseable", func() {
		var err error
		snapshotData, err = support.ParseSnapshotData()
		Expect(err).NotTo(HaveOccurred())
		support.LogMap(fmt.Sprintf("Snapshot images (%d):", len(snapshotData.Images)), snapshotData.Images)
		Expect(snapshotData.Images).NotTo(BeEmpty(), "No images were detected in snapshot file")
	})

	It("snapshot.json file contains valid images", func() {
		Expect(snapshotData.Images).To(HaveEach(MatchRegexp(support.SnapshotImageDefinitionRegexp)))
	})

	It("snapshot.json file image snapshots are all unique", func() {
		snapshotHashes := support.ExtractHashes(support.GetMapValues(snapshotData.Images))
		mapped := make(map[string]int)
		for _, hash := range snapshotHashes {
			_, exist := mapped[hash]
			if exist {
				mapped[hash]++
			} else {
				mapped[hash] = 1
			}
		}
		Expect(mapped).To(HaveEach(1))
		Expect(len(snapshotData.Images)).To(BeNumerically("==", len(mapped)))
	})
})
