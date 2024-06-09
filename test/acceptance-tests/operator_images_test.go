package acceptance_tests

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
)

var _ = Describe("Trusted Artifact Signer Operator", Ordered, func() {

	var (
		snapshotImages support.SnapshotMap
		operatorImages support.OperatorMap
		operator       string
	)

	It("get and parse snapshot.json file", func() {
		content, err := support.GetFileContent(support.GetEnvOrDefault(support.EnvReleasesSnapshotFile, support.DefaultReleasesSnapshotFile))
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(BeEmpty())

		err = json.Unmarshal([]byte(content), &snapshotImages)
		Expect(err).NotTo(HaveOccurred())
	})

	It("get operator image", func() {
		operator = snapshotImages[support.OperatorImageKey]
		Expect(operator).NotTo(BeEmpty())
		log.Printf("Using %s\n", operator)
	})

	It("get all TAS images used by this operator", func() {
		helpLogs, err := support.RunImage(operator, []string{"-h"})
		Expect(err).NotTo(HaveOccurred())
		operatorImages = support.ParseOperatorImages(helpLogs)
		log.Printf("Found %d operator TAS images\n", len(operatorImages))
	})

	It("operator images are all valid", func() {
		Expect(operatorImages).NotTo(BeEmpty())
		Expect(operatorImages).NotTo(ContainElement(BeEmpty()))
		Expect(operatorImages).To(HaveEach(MatchRegexp(support.ImageDefinitionRegexp)))
	})

	It("all image hashes are also defined in releases snapshot", func() {
		operatorHashes := support.ExtractHashes(support.GetMapValues(operatorImages))
		snapshotHashes := support.ExtractHashes(support.GetMapValues(snapshotImages))
		Expect(snapshotHashes).To(ContainElements(operatorHashes))
	})

	It("image hashes are all unique", func() {
		operatorHashes := support.ExtractHashes(support.GetMapValues(operatorImages))
		mapped := make(map[string]int)
		for _, hash := range operatorHashes {
			if _, ok := mapped[hash]; ok {
				mapped[hash]++
			} else {
				mapped[hash] = 1
			}
		}
		Expect(mapped).To(HaveEach(1))
		Expect(len(operatorImages) == len(mapped)).To(BeTrue())
	})
})
