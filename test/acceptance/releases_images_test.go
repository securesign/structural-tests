package acceptance

import (
	"fmt"

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

		// Check that all images have 'architecture' label with value 'x86_64'
		for _, imageData := range imageDataList {
			Expect(imageData.Labels).To(HaveKey("architecture"),
				fmt.Sprintf("Image %s is missing 'architecture' label", imageData.Image))
			Expect(imageData.Labels["architecture"]).To(Equal("x86_64"),
				fmt.Sprintf("Image %s has incorrect architecture label: expected 'x86_64', got '%s'",
					imageData.Image, imageData.Labels["architecture"]))
		}

		// Check that all images have 'build-date' label which is not empty
		for _, imageData := range imageDataList {
			Expect(imageData.Labels).To(HaveKey("build-date"),
				fmt.Sprintf("Image %s is missing 'build-date' label", imageData.Image))
			Expect(imageData.Labels["build-date"]).NotTo(BeEmpty(),
				fmt.Sprintf("Image %s has empty 'build-date' label", imageData.Image))
		}

		// Check that all images have 'short-commit' label which is not empty
		for _, imageData := range imageDataList {
			Expect(imageData.Labels).To(HaveKey("short-commit"),
				fmt.Sprintf("Image %s is missing 'short-commit' label", imageData.Image))
			Expect(imageData.Labels["short-commit"]).NotTo(BeEmpty(),
				fmt.Sprintf("Image %s has empty 'short-commit' label", imageData.Image))
		}
	})
})
