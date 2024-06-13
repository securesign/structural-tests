package acceptance_tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
	"regexp"
)

var _ = Describe("Trusted Artifact Signer Operator", Ordered, func() {

	var (
		snapshotImages support.SnapshotMap
		operatorImages support.OperatorMap
		operator       string
	)

	It("get and parse snapshot.json file", func() {
		var err error
		snapshotImages, err = support.ParseSnapshotImages()
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotImages).NotTo(BeEmpty())
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
		support.LogMap(fmt.Sprintf("Operator TAS images (%d):", len(operatorImages)), operatorImages)
		Expect(operatorImages).NotTo(BeEmpty())
	})

	It("operator images are all valid", func() {
		Expect(support.GetMapKeys(operatorImages)).To(ContainElements(support.MandatoryOperatorImageKeys))
		Expect(len(operatorImages)).To(BeNumerically("==", len(support.MandatoryOperatorImageKeys)))
		Expect(operatorImages).To(HaveEach(MatchRegexp(support.OperatorImageDefinitionRegexp)))
	})

	It("all image hashes are also defined in releases snapshot", func() {
		mapped := make(map[string]string)
		for _, imageKey := range support.MandatoryOperatorImageKeys {
			oSha := support.ExtractHash(operatorImages[imageKey])
			sSha := support.ExtractHash(snapshotImages[imageKey])
			if oSha == sSha {
				mapped[imageKey] = "match"
			} else {
				mapped[imageKey] = "DIFFERENT HASHES"
			}
		}
		Expect(mapped).To(HaveEach("match"))
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

	It("operator-bundle use the right operator", func() {
		fileContent, err := support.RunImage(snapshotImages[support.OperatorBundleImageKey], []string{"cat", support.OperatorBundleClusterserviceversionFile})
		Expect(err).NotTo(HaveOccurred())
		Expect(fileContent).NotTo(BeEmpty())

		operatorHash := support.ExtractHash(snapshotImages[support.OperatorImageKey])
		re := regexp.MustCompile(`(\w+:\s*[\w./-]+operator[\w-]*@sha256:` + operatorHash + `)`)
		matches := re.FindAllString(fileContent, -1)
		Expect(matches).NotTo(BeEmpty())
		support.LogArray("Operator images found in operator-bundle:", matches)
	})
})
