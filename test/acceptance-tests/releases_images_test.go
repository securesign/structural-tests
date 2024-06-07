package acceptance_tests

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
	"os"
	"path/filepath"
)

var _ = Describe("Trusted Artifact Signer Releases", Ordered, func() {

	var (
		err            error
		snapshotImages support.SnapshotMap
		releasesDir    string
	)

	It("preparing repository folder", func() {
		releasesDir, err = support.GetReleasesProjectPath()
		Expect(err).NotTo(HaveOccurred())
	})

	It("snapshot.json file exist and is parseable", func() {
		snapshotFileFolder := support.GetEnvOrDefault(support.EnvReleasesSnapshotFolder, support.ReleasesSnapshotDefFolder)
		snapshotFilePath := filepath.Join(releasesDir, fmt.Sprintf("/%s/snapshot.json", snapshotFileFolder))
		log.Printf("Reading %s\n", snapshotFilePath)
		content, err := os.ReadFile(snapshotFilePath)
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(content, &snapshotImages)
		Expect(err).NotTo(HaveOccurred())
		log.Printf("Found %d snapshot images\n", len(snapshotImages))
	})

	It("snapshot.json file contains valid image definitions", func() {
		Expect(snapshotImages).NotTo(BeEmpty())
		Expect(snapshotImages).NotTo(ContainElement(BeEmpty()))
		Expect(snapshotImages).To(HaveEach(MatchRegexp(support.ImageDefinitionRegexp)))
	})

	It("snapshot.json file image definitions are all unique", func() {
		existingImages := make(map[string]struct{})
		for _, image := range support.GetMapValues(snapshotImages) {
			if _, ok := existingImages[image]; ok {
				Fail("Not unique image: " + image)
			}
			existingImages[image] = struct{}{}
		}
	})

	It("json and yaml files are all valid", func() {
		err := support.ValidateAllYamlAndJsonFiles(releasesDir)
		Expect(err).ToNot(HaveOccurred())
	})
})
