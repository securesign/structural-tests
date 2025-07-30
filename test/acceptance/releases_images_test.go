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
		var imagesData []support.ImageData
		allLabelKeys := make(map[string]int)
		for _, image := range snapshotData.Images {
			labels, err := support.InspectImageForLabels(image)
			Expect(err).NotTo(HaveOccurred())
			Expect(labels).NotTo(BeEmpty())
			var iData support.ImageData
			iData.Image = image
			iData.Labels = labels

			// calculate counts of labels
			imagesData = append(imagesData, iData)
			for key := range labels {
				_, exist := allLabelKeys[key]
				if exist {
					allLabelKeys[key]++
				} else {
					allLabelKeys[key] = 1
				}
			}
		}

		sortedKeys := support.GetMapKeysSorted(allLabelKeys)
		support.LogMapByProvidedKeys(fmt.Sprintf("Labels counts out of max %d", len(imagesData)), allLabelKeys, sortedKeys)

		for _, key := range sortedKeys {
			fmt.Printf("%s:\n", key)
			for _, img := range imagesData {
				fmt.Printf("    [%-120s] %s\n", img.Image, img.Labels[key])
			}
		}
	})
})
