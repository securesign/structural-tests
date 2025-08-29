package acceptance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

var ErrNotFoundInRegistry = errors.New("not found in registry")

var _ = Describe("Policy Controller Operator", Ordered, func() {

	var (
		snapshotData        support.SnapshotData
		repositories        *support.RepositoryList
		operatorPcoImages   support.OperatorMap
		operatorOtherImages support.OperatorMap
		operator            string
	)

	BeforeAll(func() {
		var err error
		snapshotData, err = support.ParseSnapshotData()
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotData.Images).NotTo(BeEmpty(), "No images were detected in snapshot file")

		repositories, err = support.LoadRepositoryList()
		Expect(err).NotTo(HaveOccurred())
		Expect(repositories.Data).NotTo(BeEmpty(), "No images were detected in repositories file")
	})

	It("get operator image", func() {
		operator = snapshotData.Images[support.PolicyControllerOperatorImageKey]
		Expect(operator).NotTo(BeEmpty(), "Operator image not detected in snapshot file")
		log.Printf("Using %s\n", operator)
	})

	It("get all PCO images used by this operator", func() {
		valuesFile, err := support.RunImage(operator, []string{"cat"}, []string{"helm-charts/policy-controller-operator/values.yaml"})
		Expect(err).NotTo(HaveOccurred())

		operatorPcoImages, operatorOtherImages = support.ParsePCOperatorImages(valuesFile)
		Expect(operatorPcoImages).NotTo(BeEmpty())
		Expect(operatorOtherImages).NotTo(BeEmpty())
		support.LogMap(fmt.Sprintf("Operator Pco images (%d):", len(operatorPcoImages)), operatorPcoImages)
		support.LogMap(fmt.Sprintf("Operator other images (%d):", len(operatorOtherImages)), operatorOtherImages)
	})

	It("operator images are listed in registry.redhat.io", func() {
		var errs []error
		for _, image := range operatorPcoImages {
			if repositories.FindByImage(image) == nil {
				errs = append(errs, fmt.Errorf("%w: %s", ErrNotFoundInRegistry, image))
			}
		}
		Expect(errs).To(BeEmpty())
	})

	It("operator PCO images are all valid", func() {
		Expect(support.GetMapKeys(operatorPcoImages)).To(ContainElements(support.MandatoryPcoOperatorImageKeys()))
		Expect(len(operatorPcoImages)).To(BeNumerically("==", len(support.MandatoryPcoOperatorImageKeys())))
		Expect(operatorPcoImages).To(HaveEach(MatchRegexp(support.TasImageDefinitionRegexp)))
	})

	It("operator other images are all valid", func() {
		Expect(support.GetMapKeys(operatorOtherImages)).To(ContainElements(support.OtherPCOOperatorImageKeys()))
		Expect(len(operatorOtherImages)).To(BeNumerically("==", len(support.OtherPCOOperatorImageKeys())))
		Expect(operatorOtherImages).To(HaveEach(MatchRegexp(support.OtherImageDefinitionRegexp)))
	})

	It("all image hashes are also defined in releases snapshot", func() {
		mapped := make(map[string]string)
		for _, imageKey := range support.MandatoryPcoOperatorImageKeys() {
			oSha := support.ExtractHash(operatorPcoImages[imageKey])
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
		operatorHashes := support.ExtractHashes(support.GetMapValues(operatorPcoImages))
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
		Expect(operatorPcoImages).To(HaveLen(len(mapped)))
	})

	It("operator-bundle use the right operator", func() {
		dir, err := os.MkdirTemp("", "bundle")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(dir)

		Expect(support.FileFromImage(
			context.Background(),
			snapshotData.Images[support.PolicyControllerOperatorBundleImageKey],
			support.PolicyControllerOperatorBundleClusterServiceVersionPath, dir),
		).To(Succeed())
		fileContent, err := os.ReadFile(filepath.Join(dir, support.PolicyControllerOperatorBundleClusterServiceVersionFile))
		Expect(err).NotTo(HaveOccurred())
		Expect(fileContent).NotTo(BeEmpty())

		operatorHash := support.ExtractHash(snapshotData.Images[support.PolicyControllerOperatorImageKey])
		re := regexp.MustCompile(`(\w+:\s*[\w./-]+operator[\w-]*@sha256:` + operatorHash + `)`)
		matches := re.FindAllString(string(fileContent), -1)
		Expect(matches).NotTo(BeEmpty())
		support.LogArray("Operator images found in operator-bundle:", matches)
	})
})
