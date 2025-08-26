package acceptance

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

var _ = Describe("Trusted Artifact Signer Releases", Ordered, func() {

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

	It("operator and operator bundle have both the same git reference", func() {
		operatorReference, err := support.GetImageLabel(snapshotData.Images[support.OperatorImageKey], "vcs-ref")
		Expect(err).NotTo(HaveOccurred())
		Expect(operatorReference).NotTo(BeEmpty())
		operatorBundleReference, err := support.GetImageLabel(snapshotData.Images[support.OperatorBundleImageKey], "vcs-ref")
		Expect(err).NotTo(HaveOccurred())
		Expect(operatorBundleReference).NotTo(BeEmpty())
		Expect(operatorReference).To(Equal(operatorBundleReference))
	})

	It("snapshot.json images have correct labels", func() {
		var imageDataList []support.ImageData
		imageLabelsErrors := make(map[string][]string)

		// Collect all images and their labels
		for imageName, imageDefinition := range snapshotData.Images {
			labels, err := support.InspectImageForLabels(imageDefinition)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed to inspect labels for image %s (%s)", imageName, imageDefinition))

			imageData := support.ImageData{
				Image:  imageDefinition,
				Labels: labels,
			}
			imageDataList = append(imageDataList, imageData)
		}

		// Check that all images have required labels with correct values
		requiredLabels := support.RequiredImageLabels()
		for _, imageData := range imageDataList {
			var currentImageErrors []string

			for labelName, expectedValue := range requiredLabels {
				if _, exists := imageData.Labels[labelName]; !exists {
					currentImageErrors = append(currentImageErrors,
						fmt.Sprintf("  %s: missing", labelName))
					continue
				}

				if expectedValue != "" { // Specific value expected
					if imageData.Labels[labelName] != expectedValue {
						currentImageErrors = append(currentImageErrors,
							fmt.Sprintf("  %s: %s, expected: %s",
								labelName, imageData.Labels[labelName], expectedValue))
					}
				} else { // Label must not be empty
					if imageData.Labels[labelName] == "" {
						currentImageErrors = append(currentImageErrors,
							fmt.Sprintf("  %s: missing", labelName))
					}
				}
			}

			if len(currentImageErrors) > 0 {
				imageLabelsErrors[imageData.Image] = currentImageErrors
			}
		}

		// Format errors in a human-readable way
		if len(imageLabelsErrors) > 0 {
			var errorReport strings.Builder
			for image, errors := range imageLabelsErrors {
				errorReport.WriteString(image)
				errorReport.WriteString(":\n")
				for _, error := range errors {
					errorReport.WriteString(error)
					errorReport.WriteString("\n")
				}
			}
			Fail("Label validation errors found:\n" + errorReport.String())
		}
	})
})
