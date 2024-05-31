package acceptance_tests

import (
	"encoding/json"
	"fmt"
	gitAuth "github.com/go-git/go-git/v5/plumbing/transport/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("Trusted Artifact Signer Releases", Ordered, func() {

	var (
		err              error
		snapshotJsonData map[string]interface{}
		snapshotImages   []string
		releasesDir      string
	)

	It("cloning repository", func() {
		releasesBranch := support.GetEnvOrDefault(support.EnvReleasesRepoBranch, support.ReleasesRepoDefBranch)
		githubUsername := support.GetEnv(support.EnvTestGithubUser)
		githubToken := support.GetEnvOrDefaultSecret(support.EnvTestGithubToken, "")
		releasesDir, _, err = support.GitCloneWithAuth(support.ReleasesRepo, releasesBranch,
			&gitAuth.BasicAuth{
				Username: githubUsername,
				Password: githubToken,
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("snapshot.json file exist and is parseable", func() {
		snapshotFileFolder := support.GetEnvOrDefault(support.EnvReleasesSnapshotFolder, support.ReleasesSnapshotDefFolder)
		snapshotFilePath := filepath.Join(releasesDir, fmt.Sprintf("/%s/snapshot.json", snapshotFileFolder))
		log.Printf("Reading %s\n", snapshotFilePath)
		content, err := os.ReadFile(snapshotFilePath)
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(content, &snapshotJsonData)
		Expect(err).NotTo(HaveOccurred())
	})

	It("snapshot.json file contains valid image definitions", func() {
		snapshotImages = support.GetImageDefinitionsFromJson(snapshotJsonData)
		log.Printf("Found %d images: \n%v", len(snapshotImages), strings.Join(snapshotImages, "\n"))
		Expect(snapshotImages).NotTo(BeEmpty())
	})

	It("snapshot.json file image definitions are all unique", func() {
		existingImages := make(map[string]struct{})
		for _, image := range snapshotImages {
			if _, ok := existingImages[image]; ok {
				Fail("Not unique image: " + image)
			}
			existingImages[image] = struct{}{}
		}
	})

	It("snapshot.json file image definitions have hashes as tags", func() {
		for _, image := range snapshotImages {
			parts := strings.Split(image, "@sha256:")
			Expect(parts).To(HaveLen(2), fmt.Sprintf("Not a hash tag for %s", image))
			Expect(support.HasValidSHA(parts[1])).To(BeTrue(), fmt.Sprintf("Image has invalid hash for %s", image))
		}
	})

	It("operator core image hashes are all also present in snapshot.json file", func() {
		operatorBranch := support.GetEnvOrDefault(support.EnvOperatorRepoBranch, support.OperatorRepoDefBranch)
		operatorImagesFileContent, err := support.DownloadFileContent(fmt.Sprintf(support.OperatorImagesFile, operatorBranch), "")
		Expect(err).NotTo(HaveOccurred())
		operatorImages, err := support.GetImageDefinitionsFromConstants(operatorImagesFileContent)
		Expect(err).NotTo(HaveOccurred())
		log.Printf("Found %d images in operator (images.go): \n%v", len(operatorImages), strings.Join(operatorImages, "\n"))

		operatorSHAs, err := support.ExtractImageHashes(operatorImages, "registry.redhat.io/rhtas")
		Expect(err).NotTo(HaveOccurred())
		snapshotSHAs, err := support.ExtractImageHashes(snapshotImages, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotSHAs).To(ContainElements(operatorSHAs))
	})

	It("json and yaml files are all valid", func() {
		err := support.ValidateAllYamlAndJsonFiles(releasesDir)
		Expect(err).ToNot(HaveOccurred())
	})
})
