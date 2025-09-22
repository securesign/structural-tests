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

var _ = Describe("Model Validation Operator", Ordered, func() {

	var (
		snapshotData      support.SnapshotData
		repositories      *support.RepositoryList
		operatorMvoImages support.OperatorMap
		operator          string
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
		operator = snapshotData.Images[support.ModelValidationOperatorImageKey]
		Expect(operator).NotTo(BeEmpty(), "Operator image not detected in snapshot file")
		log.Printf("Using %s\n", operator)
	})

	It("get all MV images used by this operator", func() {
		valuesFile, err := support.RunImage(operator, []string{}, []string{"-h"})
		Expect(err).NotTo(HaveOccurred())

		operatorMvoImages = support.ParseMVOperatorImages(valuesFile)

		Expect(operatorMvoImages).NotTo(BeEmpty())
		support.LogMap(fmt.Sprintf("Operator MVO images (%d):", len(operatorMvoImages)), operatorMvoImages)
	})

	It("operator images are listed in registry.redhat.io", func() {
		var errs []error

		for _, image := range operatorMvoImages {
			if repositories.FindByImage(image) == nil {
				errs = append(errs, fmt.Errorf("%w: %s", ErrNotFoundInRegistry, image))
			}
		}
		Expect(errs).To(BeEmpty())
	})

	It("operator MVO images are all valid", func() {

		Expect(support.GetMapKeys(operatorMvoImages)).To(ContainElements(support.MandatoryMvoOperatorImageKeys()))
		Expect(len(operatorMvoImages)).To(BeNumerically("==", len(support.MandatoryMvoOperatorImageKeys())))
		Expect(operatorMvoImages).To(HaveEach(MatchRegexp(support.TasImageDefinitionRegexp)))
	})

	It("all image hashes are also defined in releases snapshot", func() {

		mapped := make(map[string]string)
		for _, imageKey := range support.MandatoryMvoOperatorImageKeys() {
			oSha := support.ExtractHash(operatorMvoImages[imageKey])
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

		operatorHashes := support.ExtractHashes(support.GetMapValues(operatorMvoImages))
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
		Expect(operatorMvoImages).To(HaveLen(len(mapped)))
	})

	It("operator-bundle use the right operator", func() {
		dir, err := os.MkdirTemp("", "bundle")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(dir)

		Expect(support.FileFromImage(
			context.Background(),
			snapshotData.Images[support.ModelValidationOperatorBundleImageKey],
			support.ModelValidationOperatorBundleClusterServiceVersionPath, dir),
		).To(Succeed())
		fileContent, err := os.ReadFile(filepath.Join(dir, support.ModelValidationOperatorBundleClusterServiceVersionFile))
		Expect(err).NotTo(HaveOccurred())
		Expect(fileContent).NotTo(BeEmpty())

		operatorHash := support.ExtractHash(snapshotData.Images[support.ModelValidationOperatorImageKey])
		re := regexp.MustCompile(`(\w+:\s*[\w./-]+operator[\w-]*@sha256:` + operatorHash + `)`)
		matches := re.FindAllString(string(fileContent), -1)
		Expect(matches).NotTo(BeEmpty())
		support.LogArray("Operator images found in operator-bundle:", matches)
	})
})
