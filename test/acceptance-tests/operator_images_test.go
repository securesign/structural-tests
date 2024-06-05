package acceptance_tests

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
)

var _ = Describe("Trusted Artifact Signer Operator", func() {

	var (
		snapshotJsonData        support.Snapshot
		operatorImageDefinition string
		operatorImages          map[string]string
	)

	It("get and parse snapshot.json file", func() {
		releasesBranch := support.GetEnvOrDefault(support.EnvReleasesRepoBranch, support.ReleasesRepoDefBranch)
		snapshotFileFolder := support.GetEnvOrDefault(support.EnvReleasesSnapshotFolder, support.ReleasesSnapshotDefFolder)
		snapshotFile := fmt.Sprintf(support.ReleasesSnapshotFile, releasesBranch, snapshotFileFolder)
		githubToken := support.GetEnvOrDefaultSecret(support.EnvTestGithubToken, "")
		content, err := support.DownloadFileContent(snapshotFile, githubToken)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(BeEmpty())

		err = json.Unmarshal([]byte(content), &snapshotJsonData)
		Expect(err).NotTo(HaveOccurred())
	})

	It("get operator image", func() {
		operatorImageDefinition = snapshotJsonData.Operator.RhtasOperator
		Expect(operatorImageDefinition).NotTo(BeEmpty())
		log.Printf("Using %s", operatorImageDefinition)
	})

	It("get all images used by this operator", func() {
		helpLogs, err := support.RunImage(operatorImageDefinition, []string{"-h"})
		Expect(err).NotTo(HaveOccurred())
		operatorImages = support.ParseOperatorImages(helpLogs)
		Expect(operatorImages).NotTo(BeEmpty())
		log.Printf("Found %d operator images: %v\n", len(operatorImages), operatorImages)
	})

	It("all images are defined in releases snapshot", func() {
		notMatching := make(map[string]string)
		for opKey, operatorImage := range operatorImages {
			if opKey == "client-server-image" || opKey == "trillian-netcat-image" {
				continue // these are not TAS images
			}
			snapshotImage := support.GetCorrespondingSnapshotImage(opKey, snapshotJsonData)
			if !support.ImageHashesIdentical(operatorImage, snapshotImage) {
				notMatching[operatorImage] = snapshotImage
			}
		}
		if len(notMatching) > 0 {
			log.Printf("Found %d not matching operator/snapshot image hashes:\n", len(notMatching))
			for op, sn := range notMatching {
				log.Printf("%s\n", op)
				log.Printf("  %s\n", sn)
			}
			Fail(fmt.Sprintf("%d operator/snapshot image hashes do not match", len(notMatching)))
		}
	})
})
