package acceptance

import (
	"fmt"
	"log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

var _ = Describe("Trusted Artifact Signer Ansible", Ordered, func() {

	var (
		snapshotImages     support.SnapshotMap
		repositories       *support.RepositoryList
		ansibleAllImages   support.AnsibleMap
		ansibleTasImages   support.AnsibleMap
		ansibleOtherImages support.AnsibleMap
	)

	It("get and parse snapshot file", func() {
		var err error
		snapshotImages, err = support.ParseSnapshotImages()
		support.LogMap(fmt.Sprintf("Snapshot images (%d):", len(snapshotImages)), snapshotImages)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotImages).NotTo(BeEmpty(), "No images were detected in snapshot file")

		repositories, err = support.LoadRepositoryList()
		Expect(err).NotTo(HaveOccurred())
		Expect(repositories.Data).NotTo(BeEmpty(), "No images were detected in repositories file")
	})

	It("load and parse ansible definition file", func() {
		var err error
		forcedAnsibleFileName := support.GetEnv(support.EnvAnsibleImagesFile)
		if forcedAnsibleFileName == "" {
			// standard way - use ansible definition file name from releases snapshot.json
			ansibleCollectionUrl := snapshotImages[support.AnsibleCollectionKey]
			Expect(ansibleCollectionUrl).NotTo(BeEmpty())
			log.Printf("Using %s URL from snapshot.json file\n", ansibleCollectionUrl)

			fileContent, err := support.LoadFileFromZip(ansibleCollectionUrl, support.AnsibleCollectionSnapshotFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileContent).NotTo(BeEmpty(), "Ansible definition file seems to be empty")
			ansibleAllImages, err = support.MapAnsibleImages(fileContent)
		} else {
			// use ansible definition file defined via env variable
			ansibleAllImages, err = support.ParseAnsibleImages(forcedAnsibleFileName)
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(ansibleAllImages).NotTo(BeEmpty())
	})

	It("get and parse ansible image definition file", func() {
		ansibleTasImages, ansibleOtherImages = support.SplitMap(ansibleAllImages, support.AnsibleTasImageKeys())
		Expect(ansibleTasImages).NotTo(BeEmpty())
		Expect(ansibleOtherImages).NotTo(BeEmpty())
		support.LogMap(fmt.Sprintf("Ansible TAS images (%d):", len(ansibleTasImages)), ansibleTasImages)
		support.LogMap(fmt.Sprintf("Ansible other images (%d):", len(ansibleOtherImages)), ansibleOtherImages)
	})

	It("ansible TAS images are listed in registry.redhat.io", func() {
		var errs []error
		for _, ansibleImage := range ansibleTasImages {
			if repositories.FindByImage(ansibleImage) == nil {
				errs = append(errs, fmt.Errorf("%w: %s", ErrNotFoundInRegistry, ansibleImage))
			}
		}
		Expect(errs).To(BeEmpty())
	})

	It("ansible TAS images are all valid", func() {
		Expect(support.GetMapKeys(ansibleTasImages)).To(ContainElements(support.AnsibleTasImageKeys()))
		Expect(len(ansibleTasImages)).To(BeNumerically("==", len(support.AnsibleTasImageKeys())))
		Expect(ansibleTasImages).To(HaveEach(MatchRegexp(support.TasImageDefinitionRegexp)))
	})

	It("ansible other images are all valid", func() {
		Expect(support.GetMapKeys(ansibleOtherImages)).To(ContainElements(support.AnsibleOtherImageKeys()))
		Expect(len(ansibleOtherImages)).To(BeNumerically("==", len(support.AnsibleOtherImageKeys())))
		Expect(ansibleOtherImages).To(HaveEach(MatchRegexp(support.OtherImageDefinitionRegexp)))
	})

	It("all ansible TAS image hashes are also defined in releases snapshot", func() {
		mapped := make(map[string]string)
		for _, imageKey := range support.AnsibleTasImageKeys() {

			// skip, while ansible uses older tuf image
			if imageKey == "tas_single_node_tuf_image" {
				log.Printf("Ansible uses differet TUF image - skipping")
				log.Printf("  Ansible:  %s", ansibleTasImages[imageKey])
				log.Printf("  Snapshot: %s", snapshotImages[support.ConvertAnsibleImageKey(imageKey)])
				continue
			}

			aSha := support.ExtractHash(ansibleTasImages[imageKey])
			if _, keyExist := snapshotImages[support.ConvertAnsibleImageKey(imageKey)]; !keyExist {
				mapped[imageKey] = "MISSING"
				continue
			}
			sSha := support.ExtractHash(snapshotImages[support.ConvertAnsibleImageKey(imageKey)])
			if aSha == sSha {
				mapped[imageKey] = "match"
			} else {
				mapped[imageKey] = "DIFFERENT HASHES"
			}
		}
		Expect(mapped).To(HaveEach("match"), "Ansible images are missing or have different hashes in snapshot file")
	})

	It("image hashes are all unique", func() {
		aImageHashes := support.ExtractHashes(support.GetMapValues(ansibleTasImages))
		hashesCounts := make(map[string]int)
		for _, hash := range aImageHashes {
			_, exist := hashesCounts[hash]
			if exist {
				hashesCounts[hash]++
			} else {
				hashesCounts[hash] = 1
			}
		}
		Expect(hashesCounts).To(HaveEach(1))
		Expect(ansibleTasImages).To(HaveLen(len(hashesCounts)))
	})

})
