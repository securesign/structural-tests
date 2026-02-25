package acceptance

import (
	"fmt"
	"log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"golang.org/x/mod/semver"
)

var _ = Describe("Trusted Artifact Signer Ansible", Ordered, func() {

	var (
		snapshotData support.SnapshotData
		repositories *support.RepositoryList

		ansibleFileContent   []byte
		ansibleCollectionURL string

		ansibleTasImages   support.AnsibleMap
		ansibleOtherImages support.AnsibleMap

		ansibleTasKeys   []string
		ansibleOtherKeys []string
	)

	BeforeAll(func() {
		By("get and parse snapshot file")
		var err error
		snapshotData, err = support.ParseSnapshotData()
		support.LogMap(fmt.Sprintf("Snapshot images (%d):", len(snapshotData.Images)), snapshotData.Images)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotData.Images).NotTo(BeEmpty(), "No images were detected in snapshot file")

		repositories, err = support.LoadRepositoryList()
		Expect(err).NotTo(HaveOccurred())
		Expect(repositories.Data).NotTo(BeEmpty(), "No images were detected in repositories file")

		By("resolve ansible collection URL")
		ansibleCollectionURL = support.GetEnv(support.EnvAnsibleImagesFile)
		if ansibleCollectionURL == "" {
			support.LogAvailableAnsibleArtifacts()
			// standard way - use ansible definition file path from releases snapshot.json file
			snapshotAnsibleURL := snapshotData.Others[support.AnsibleCollectionKey]
			if snapshotAnsibleURL != "" {
				log.Printf("Using %s URL from snapshot.json file\n", snapshotAnsibleURL)
				ansibleCollectionURL, err = support.MapAnsibleZipFileURL(snapshotAnsibleURL)
				Expect(err).NotTo(HaveOccurred())
			}
		}

		By("check supported version")
		version := support.GetEnv(support.EnvVersion)
		if semver.Compare("v"+version, "v1.2.0") < 0 && ansibleCollectionURL == "" {
			Skip("Ansible is optional for " + version)
		}

		By("load ansible image key lists from config")
		ansibleTasKeys, ansibleOtherKeys, err = support.GetAnsibleImageKeysFromConfig(defaults)
		Expect(err).NotTo(HaveOccurred())
	})

	It("load ansible definition file", func() {
		var err error
		Expect(ansibleCollectionURL).NotTo(BeEmpty())
		ansibleFileContent, err = support.LoadAnsibleCollectionSnapshotFile(ansibleCollectionURL, support.AnsibleCollectionSnapshotFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(ansibleFileContent).NotTo(BeEmpty(), "Ansible definition file seems to be empty")
	})

	It("get and parse ansible images definition file", func() {
		ansibleAllImages, err := support.MapAnsibleImages(ansibleFileContent)
		Expect(err).NotTo(HaveOccurred())
		Expect(ansibleAllImages).NotTo(BeEmpty())
		ansibleTasImages, ansibleOtherImages = support.SplitMap(ansibleAllImages, ansibleTasKeys)
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
		Expect(support.GetMapKeys(ansibleTasImages)).To(ContainElements(ansibleTasKeys))
		Expect(len(ansibleTasImages)).To(BeNumerically("==", len(ansibleTasKeys)))
		Expect(ansibleTasImages).To(HaveEach(MatchRegexp(support.TasImageDefinitionRegexp)))
	})

	It("ansible other images are all valid", func() {
		Expect(support.GetMapKeys(ansibleOtherImages)).To(ContainElements(ansibleOtherKeys))
		Expect(len(ansibleOtherImages)).To(BeNumerically("==", len(ansibleOtherKeys)))
		Expect(ansibleOtherImages).To(HaveEach(MatchRegexp(support.OtherImageDefinitionRegexp)))
	})

	It("all ansible TAS image hashes are also defined in releases snapshot", func() {
		mapped := make(map[string]string)
		for _, imageKey := range ansibleTasKeys {
			aSha := support.ExtractHash(ansibleTasImages[imageKey])
			if _, keyExist := snapshotData.Images[support.ConvertAnsibleImageKey(imageKey)]; !keyExist {
				mapped[imageKey] = "MISSING"
				continue
			}
			sSha := support.ExtractHash(snapshotData.Images[support.ConvertAnsibleImageKey(imageKey)])
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
