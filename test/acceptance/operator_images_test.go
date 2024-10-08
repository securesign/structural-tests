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
		snapshotImages      support.SnapshotMap
		repositories        *support.RepositoryList
		operatorTasImages   support.OperatorMap
		operatorOtherImages support.OperatorMap
		operator            string
	)

	It("get and parse snapshot.json file", func() {
		var err error
		snapshotImages, err = support.ParseSnapshotImages()
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotImages).NotTo(BeEmpty())

		repositories, err = support.LoadRepositoryList()
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
		Expect(operatorTasImages).To(HaveEach(MatchRegexp(support.OperatorTasImageDefinitionRegexp)))
	})

	It("operator other images are all valid", func() {
		Expect(support.GetMapKeys(operatorOtherImages)).To(ContainElements(support.OtherOperatorImageKeys()))
		Expect(len(operatorOtherImages)).To(BeNumerically("==", len(support.OtherOperatorImageKeys())))
		Expect(operatorOtherImages).To(HaveEach(MatchRegexp(support.OtherOperatorImageDefinitionRegexp)))
	})

	It("all image hashes are also defined in releases snapshot", func() {
		mapped := make(map[string]string)
		for _, imageKey := range support.MandatoryTasOperatorImageKeys() {
			oSha := support.ExtractHash(operatorTasImages[imageKey])
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
		fileContent, err := support.RunImage(snapshotImages[support.OperatorBundleImageKey], []string{"cat", support.OperatorBundleClusterServiceVersionFile})
		Expect(err).NotTo(HaveOccurred())
		Expect(fileContent).NotTo(BeEmpty())

		operatorHash := support.ExtractHash(snapshotImages[support.OperatorImageKey])
		re := regexp.MustCompile(`(\w+:\s*[\w./-]+operator[\w-]*@sha256:` + operatorHash + `)`)
		matches := re.FindAllString(fileContent, -1)
		Expect(matches).NotTo(BeEmpty())
		support.LogArray("Operator images found in operator-bundle:", matches)
	})
})
