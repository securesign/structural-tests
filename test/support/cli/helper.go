package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck
	. "github.com/onsi/gomega"    //nolint:stylecheck
	"github.com/securesign/structural-tests/test/support"
)

const archiveTimeout = 5 * time.Minute

// DescribeCLIStackImageTests registers only the product's configured CLI inventory.
func DescribeCLIStackImageTests(product string, defaultsData []byte) bool {
	cfg, configErr := GetCLIStackConfig(product, defaultsData)
	return Describe("CLI Stack Images", func() {
		if configErr != nil {
			It("loads valid CLI stack configuration", func() {
				Expect(configErr).NotTo(HaveOccurred())
			})
			return
		}
		if len(cfg.Images) == 0 {
			It("has a configured CLI inventory", func() {
				Skip("CLI stack checks disabled by product configuration")
			})
			return
		}
		for _, entry := range cfg.Images {
			Describe(entry.ImageKey, Ordered, func() {
				var image string
				It("image present in snapshot", func() {
					snapshot, err := support.ParseSnapshotData()
					Expect(err).NotTo(HaveOccurred())
					image = snapshot.Images[entry.ImageKey]
					Expect(image).NotTo(BeEmpty(), "image key %q not found in snapshot", entry.ImageKey)
				})
				for _, binary := range entry.Binaries {
					It(fmt.Sprintf("contains valid %s binary for %s/%s", filepath.Base(binary.Path), binary.OS, binary.Arch), func() {
						workDir, err := os.MkdirTemp("", "cli-stack")
						Expect(err).NotTo(HaveOccurred())
						DeferCleanup(func() { Expect(os.RemoveAll(workDir)).To(Succeed()) })

						ctx, cancel := context.WithTimeout(context.Background(), archiveTimeout)
						defer cancel()
						Expect(support.FileFromImage(ctx, image, binary.Path, workDir)).To(
							Succeed(), "extracting %s from %s", binary.Path, entry.ImageKey,
						)
						executable, err := support.ExtractFirstFileFromTarGz(filepath.Join(workDir, filepath.Base(binary.Path)), workDir)
						Expect(err).NotTo(HaveOccurred())
						Expect(verifyBinaryExecutable(executable, binary.OS, binary.Arch)).To(Succeed())
					})
				}
			})
		}
	})
}
