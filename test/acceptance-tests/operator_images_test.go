package acceptance_tests

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
)

var _ = Describe("Trusted Artifact Signer Operator", func() {

	var (
		snapshotImages support.SnapshotMap
		operatorImage  string
		operatorImages map[string]string
	)

	It("get and parse snapshot.json file", func() {
		var content string
		var err error
		snapshotFile, isLocal := support.GetReleasesSnapshotFilePath()
		if isLocal {
			content, err = support.LoadFileContent(snapshotFile)
		} else {
			githubToken := support.GetEnvOrDefaultSecret(support.EnvTestGithubToken, "")
			content, err = support.DownloadFileContent(snapshotFile, githubToken)
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(BeEmpty())

		err = json.Unmarshal([]byte(content), &snapshotImages)
		Expect(err).NotTo(HaveOccurred())
	})

	It("get operator image", func() {
		operatorImage = snapshotImages[support.OperatorImageKey]
		Expect(operatorImage).NotTo(BeEmpty())
		log.Printf("Using %s\n", operatorImage)
	})

	It("get all images used by this operator", func() {
		helpLogs, err := support.RunImage(operatorImage, []string{"-h"})
		Expect(err).NotTo(HaveOccurred())
		operatorImages = support.ParseOperatorImages(helpLogs)
		log.Printf("Found %d operator images\n", len(operatorImages))
		Expect(operatorImages).NotTo(BeEmpty())
		Expect(operatorImages).NotTo(ContainElement(BeEmpty()))
		Expect(operatorImages).To(HaveEach(MatchRegexp(support.ImageDefinitionRegexp)))
	})

	It("all images are defined in releases snapshot", func() {
		operatorHashes := support.ExtractHashes(support.GetMapValues(operatorImages))
		snapshotHashes := support.ExtractHashes(support.GetMapValues(snapshotImages))
		Expect(snapshotHashes).To(ContainElements(operatorHashes))
	})
})
