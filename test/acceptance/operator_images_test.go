package acceptance

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

var ErrNotFoundInRegistry = errors.New("not found in registry")

var _ = Describe("Trusted Artifact Signer Operator", Ordered, func() {

	var (
		snapshotData        support.SnapshotData
		repositories        *support.RepositoryList
		operatorTasImages   support.OperatorMap
		operatorOtherImages support.OperatorMap
		operator            string
	)

	It("get and parse snapshot file", func() {
		var err error
		snapshotData, err = support.ParseSnapshotData()
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotData.Images).NotTo(BeEmpty(), "No images were detected in snapshot file")

		repositories, err = support.LoadRepositoryList()
		Expect(err).NotTo(HaveOccurred())
		Expect(repositories.Data).NotTo(BeEmpty(), "No images were detected in repositories file")
	})

	It("get operator image", func() {
		operator = snapshotData.Images[support.OperatorImageKey]
		Expect(operator).NotTo(BeEmpty(), "Operator image not detected in snapshot file")
		log.Printf("Using %s\n", operator)
	})

	It("get all TAS images used by this operator", func() {
		helpLogs, err := support.RunImage(operator, []string{"-h"})
		Expect(err).NotTo(HaveOccurred())

		operatorTasImages, operatorOtherImages = support.ParseOperatorImages(helpLogs)
		support.LogMap(fmt.Sprintf("Operator TAS images (%d):", len(operatorTasImages)), operatorTasImages)
		support.LogMap(fmt.Sprintf("Operator other images (%d):", len(operatorOtherImages)), operatorOtherImages)
		Expect(operatorTasImages).NotTo(BeEmpty())
		Expect(operatorOtherImages).NotTo(BeEmpty())
	})

	It("operator images are listed in registry.redhat.io", func() {
		var errs []error
		for _, image := range operatorTasImages {
			if repositories.FindByImage(image) == nil {
				errs = append(errs, fmt.Errorf("%w: %s", ErrNotFoundInRegistry, image))
			}
		}
		Expect(errs).To(BeEmpty())
	})

	It("operator TAS images are all valid", func() {
		Expect(support.GetMapKeys(operatorTasImages)).To(ContainElements(support.MandatoryTasOperatorImageKeys()))
		Expect(len(operatorTasImages)).To(BeNumerically("==", len(support.MandatoryTasOperatorImageKeys())))
		Expect(operatorTasImages).To(HaveEach(MatchRegexp(support.TasImageDefinitionRegexp)))
	})

	It("operator other images are all valid", func() {
		Expect(support.GetMapKeys(operatorOtherImages)).To(ContainElements(support.OtherOperatorImageKeys()))
		Expect(len(operatorOtherImages)).To(BeNumerically("==", len(support.OtherOperatorImageKeys())))
		Expect(operatorOtherImages).To(HaveEach(MatchRegexp(support.OtherImageDefinitionRegexp)))
	})

	It("all image hashes are also defined in releases snapshot", func() {
		mapped := make(map[string]string)
		for _, imageKey := range support.MandatoryTasOperatorImageKeys() {
			oSha := support.ExtractHash(operatorTasImages[imageKey])
			if _, keyExist := snapshotData.Images[imageKey]; !keyExist {
				mapped[imageKey] = "MISSING"
				continue
			}
			sSha := support.ExtractHash(snapshotData.Images[imageKey])
			if oSha == sSha {
				mapped[imageKey] = "match"
			} else {
				mapped[imageKey] = "DIFFERENT HASHES"
			}
		}
		Expect(mapped).To(HaveEach("match"), "Operator images are missing or have different hashes in snapshot file")
	})

	It("image hashes are all unique", func() {
		operatorHashes := support.ExtractHashes(support.GetMapValues(operatorTasImages))
		mapped := make(map[string]int)
		for _, hash := range operatorHashes {
			_, exist := mapped[hash]
			if exist {
				mapped[hash]++
			} else {
				mapped[hash] = 1
			}
		}
		Expect(mapped).To(HaveEach(1))
		Expect(operatorTasImages).To(HaveLen(len(mapped)))
	})

	It("operator-bundle use the right operator", func() {
		fileContent, err := support.RunImage(snapshotData.Images[support.OperatorBundleImageKey], []string{"cat", support.OperatorBundleClusterServiceVersionFile})
		Expect(err).NotTo(HaveOccurred())
		Expect(fileContent).NotTo(BeEmpty())

		operatorHash := support.ExtractHash(snapshotData.Images[support.OperatorImageKey])
		re := regexp.MustCompile(`(\w+:\s*[\w./-]+operator[\w-]*@sha256:` + operatorHash + `)`)
		matches := re.FindAllString(fileContent, -1)
		Expect(matches).NotTo(BeEmpty())
		support.LogArray("Operator images found in operator-bundle:", matches)
	})
})
