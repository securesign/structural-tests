package operator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck
	. "github.com/onsi/gomega"    //nolint:stylecheck
	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/pyxis"
)

func DescribeOperatorImageTests(product string, defaultsData []byte) bool {
	return Describe("Operator images", Ordered, func() {
		var (
			cfg                 OperatorConfig
			snapshotData        support.SnapshotData
			repositories        *support.RepositoryList
			operatorImage       string
			operatorTasImages   support.OperatorMap
			operatorOtherImages support.OperatorMap
		)

		BeforeAll(func() {
			var err error
			cfg, err = GetOperatorConfig(product, defaultsData)
			Expect(err).NotTo(HaveOccurred(), "failed to load operator config for product %q", product)

			snapshotData, err = support.ParseSnapshotData()
			Expect(err).NotTo(HaveOccurred())
			Expect(snapshotData.Images).NotTo(BeEmpty(), "No images were detected in snapshot file")

			repositories, err = support.LoadRepositoryList()
			Expect(err).NotTo(HaveOccurred())
			Expect(repositories.Data).NotTo(BeEmpty(), "No images were detected in repositories file")
		})

		It("get operator image", func() {
			operatorImage = snapshotData.Images[cfg.OperatorImageKey]
			Expect(operatorImage).NotTo(BeEmpty(), "Operator image not detected in snapshot file")
			log.Printf("Using %s\n", operatorImage)
		})

		It("get all images used by this operator", func() {
			output, err := support.RunImage(operatorImage, cfg.Entrypoint, []string{cfg.Entrypointcmd})
			Expect(err).NotTo(HaveOccurred())

			switch cfg.ParseFormat {
			case "values":
				operatorTasImages, operatorOtherImages = support.ParsePCOperatorImages(output)
			default:
				operatorTasImages, operatorOtherImages = support.ParseOperatorImages(output, cfg.OtherImageKeys)
			}

			support.LogMap(fmt.Sprintf("Operator TAS images (%d):", len(operatorTasImages)), operatorTasImages)
			if len(operatorOtherImages) > 0 {
				support.LogMap(fmt.Sprintf("Operator other images (%d):", len(operatorOtherImages)), operatorOtherImages)
			}
			Expect(operatorTasImages).NotTo(BeEmpty())
		})

		It("operator images are listed in registry.redhat.io", func() {
			var errs []error
			for _, image := range operatorTasImages {
				if repositories.FindByImage(image) == nil {
					errs = append(errs, fmt.Errorf("%w: %s", errors.New("not found in registry"), image))
				}
			}
			Expect(errs).To(BeEmpty())
		})

		It("operator TAS images are all valid", func() {
			Expect(support.GetMapKeys(operatorTasImages)).To(ContainElements(cfg.ImageKeys))
			Expect(len(operatorTasImages)).To(BeNumerically("==", len(cfg.ImageKeys)))
			Expect(operatorTasImages).To(HaveEach(MatchRegexp(support.TasImageDefinitionRegexp)))
		})

		if len(cfg.OtherImageKeys) > 0 {
			It("operator other images are all valid", func() {
				Expect(support.GetMapKeys(operatorOtherImages)).To(ContainElements(cfg.OtherImageKeys))
				Expect(len(operatorOtherImages)).To(BeNumerically("==", len(cfg.OtherImageKeys)))
				Expect(operatorOtherImages).To(HaveEach(MatchRegexp(support.OtherImageDefinitionRegexp)))
			})
		}

		It("all image hashes are also defined in releases snapshot", func() {
			mapped := make(map[string]string)
			for _, imageKey := range cfg.ImageKeys {
				snapshotKey := cfg.SnapshotKey(imageKey)
				oSha := support.ExtractHash(operatorTasImages[imageKey])
				if _, keyExist := snapshotData.Images[snapshotKey]; !keyExist {
					mapped[imageKey] = "MISSING"
					continue
				}
				sSha := support.ExtractHash(snapshotData.Images[snapshotKey])
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
			hashesCounts := make(map[string]int)
			for _, hash := range operatorHashes {
				_, exist := hashesCounts[hash]
				if exist {
					hashesCounts[hash]++
				} else {
					hashesCounts[hash] = 1
				}
			}
			Expect(hashesCounts).To(HaveEach(1))
			Expect(operatorTasImages).To(HaveLen(len(hashesCounts)))
		})

		if cfg.BundleImageKey != "" {
			It("operator-bundle use the right operator", func() {
				bundleImage := snapshotData.Images[cfg.BundleImageKey]
				Expect(bundleImage).NotTo(BeEmpty(), "Bundle image %q not found in snapshot", cfg.BundleImageKey)

				dir, err := os.MkdirTemp("", "bundle")
				Expect(err).NotTo(HaveOccurred())
				defer os.RemoveAll(dir)

				Expect(support.FileFromImage(
					context.Background(),
					bundleImage,
					cfg.BundleCSVPath, dir),
				).To(Succeed())

				csvFile := filepath.Base(cfg.BundleCSVPath)
				fileContent, err := os.ReadFile(filepath.Join(dir, csvFile))
				Expect(err).NotTo(HaveOccurred())
				Expect(fileContent).NotTo(BeEmpty())

				operatorHash := support.ExtractHash(snapshotData.Images[cfg.OperatorImageKey])
				re := regexp.MustCompile(`(\w+:\s*[\w./-]+operator[\w-]*@sha256:` + operatorHash + `)`)
				matches := re.FindAllString(string(fileContent), -1)
				Expect(matches).NotTo(BeEmpty())
				support.LogArray("Operator images found in operator-bundle:", matches)
			})
		}

		if len(cfg.OtherImageKeys) > 0 {
			It("other images have acceptable grades", func() {
				Expect(operatorOtherImages).NotTo(BeEmpty(), "No other images found to check grades for")
				results, err := pyxis.FetchGradesForImages(operatorOtherImages)
				Expect(err).NotTo(HaveOccurred())
				Expect(results.NotFound).To(BeEmpty(), "Some operator other images were not found in Pyxis")
				Expect(results.Grades).NotTo(BeEmpty())
				errs := pyxis.ValidateGrades(results.Grades, pyxis.GradeB, 7)
				Expect(errs).To(BeEmpty(), "Some operator other images have unacceptable grades")
			})
		}
	})
}
