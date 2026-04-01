package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

// cliStackEntry describes one CLI stack image and the binaries it must contain.
type cliStackEntry struct {
	// snapshotKey is the key in snapshot.json Images map.
	snapshotKey string
	// binaries is the list of expected binaries within the image.
	binaries []cliStackBinary
}

// cliStackBinary describes one binary tarball inside a cli-stack image.
type cliStackBinary struct {
	// pathInImage is the full path to the .tar.gz (or .exe.gz) inside the image, under /binaries/.
	pathInImage string
	// osName is the target OS for executable verification ("linux", "darwin", "windows").
	// Empty means skip executable format verification (e.g. unsupported format).
	osName string
	// arch is the target architecture ("amd64", "arm64", "ppc64le", "s390x").
	arch string
}

func tarGzPath(name, osName, arch string) string {
	return fmt.Sprintf("/binaries/%s_%s_%s.tar.gz", name, osName, arch)
}

// cliStackEntries defines all expected cli-stack images and their binaries.
// Each entry maps directly to a *-cli-stack-image key in the snapshot.
var cliStackEntries = []cliStackEntry{ //nolint:gochecknoglobals // test table
	{
		snapshotKey: "cosign-cli-stack-image",
		binaries: []cliStackBinary{
			{tarGzPath("cosign", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("cosign", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("cosign", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("cosign", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("cosign", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("cosign", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/cosign_windows_amd64.tar.gz", "windows", "amd64"},
		},
	},
	{
		snapshotKey: "gitsign-cli-stack-image",
		binaries: []cliStackBinary{
			{tarGzPath("gitsign_cli", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("gitsign_cli", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("gitsign_cli", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("gitsign_cli", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("gitsign_cli", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("gitsign_cli", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/gitsign_cli_windows_amd64.tar.gz", "windows", "amd64"},
		},
	},
	{
		snapshotKey: "rekor-cli-stack-image",
		binaries: []cliStackBinary{
			{tarGzPath("rekor_cli", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("rekor_cli", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("rekor_cli", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("rekor_cli", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("rekor_cli", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("rekor_cli", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/rekor_cli_windows_amd64.tar.gz", "windows", "amd64"},
		},
	},
	{
		snapshotKey: "fetch-tsa-certs-cli-stack-image",
		binaries: []cliStackBinary{
			{tarGzPath("fetch_tsa_certs", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("fetch_tsa_certs", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("fetch_tsa_certs", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("fetch_tsa_certs", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("fetch_tsa_certs", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("fetch_tsa_certs", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/fetch_tsa_certs_windows_amd64.tar.gz", "windows", "amd64"},
		},
	},
	{
		snapshotKey: "trillian-cli-stack-image",
		binaries: []cliStackBinary{
			{tarGzPath("createtree", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("createtree", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("createtree", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("createtree", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("createtree", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("createtree", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/createtree_windows_amd64.tar.gz", "windows", "amd64"},
			{tarGzPath("updatetree", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("updatetree", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("updatetree", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("updatetree", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("updatetree", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("updatetree", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/updatetree_windows_amd64.tar.gz", "windows", "amd64"},
		},
	},
	{
		snapshotKey: "tuftool-cli-stack-image",
		binaries: []cliStackBinary{
			// tuftool is linux/amd64 only
			{tarGzPath("tuftool", "linux", "amd64"), "linux", "amd64"},
		},
	},
	{
		snapshotKey: "conforma-cli-stack-image",
		binaries: []cliStackBinary{
			{tarGzPath("ec", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("ec", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("ec", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("ec", "linux", "s390x"), "linux", "s390x"},
			{tarGzPath("ec", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("ec", "darwin", "arm64"), "darwin", "arm64"},
			{"/binaries/ec_windows_amd64.tar.gz", "windows", "amd64"},
		},
	},
	{
		snapshotKey: "model-transparency-cli-stack-image",
		binaries: []cliStackBinary{
			// Standard variants
			{tarGzPath("model_transparency_cli", "linux", "amd64"), "linux", "amd64"},
			{tarGzPath("model_transparency_cli", "linux", "arm64"), "linux", "arm64"},
			{tarGzPath("model_transparency_cli", "linux", "ppc64le"), "linux", "ppc64le"},
			{tarGzPath("model_transparency_cli", "linux", "s390x"), "linux", "s390x"},
			// PKCS11 variants (linux only, skip exe check — same ELF binary)
			{"/binaries/model_transparency_cli_pkcs11_linux_amd64.tar.gz", "linux", "amd64"},
			{"/binaries/model_transparency_cli_pkcs11_linux_arm64.tar.gz", "linux", "arm64"},
			{"/binaries/model_transparency_cli_pkcs11_linux_ppc64le.tar.gz", "linux", "ppc64le"},
			{"/binaries/model_transparency_cli_pkcs11_linux_s390x.tar.gz", "linux", "s390x"},
			// OTel variants (linux only)
			{"/binaries/model_transparency_cli_otel_linux_amd64.tar.gz", "linux", "amd64"},
			{"/binaries/model_transparency_cli_otel_linux_arm64.tar.gz", "linux", "arm64"},
			{"/binaries/model_transparency_cli_otel_linux_ppc64le.tar.gz", "linux", "ppc64le"},
			{"/binaries/model_transparency_cli_otel_linux_s390x.tar.gz", "linux", "s390x"},
			// Cross-compiled
			{tarGzPath("model_transparency_cli", "darwin", "amd64"), "darwin", "amd64"},
			{tarGzPath("model_transparency_cli", "darwin", "arm64"), "darwin", "arm64"},
			// Windows binary is packaged as a tar inside .exe.gz
			{"/binaries/model_transparency_cli_windows_amd64.exe.gz", "windows", "amd64"},
		},
	},
}

var _ = Describe("CLI Stack Images", Ordered, func() {

	var (
		snapshotData support.SnapshotData
		tmpDir       string
	)

	BeforeAll(func() {
		var err error
		snapshotData, err = support.ParseSnapshotData()
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshotData.Images).NotTo(BeEmpty())

		tmpDir, err = os.MkdirTemp("", "cli-stack")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		if tmpDir != "" {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		}
	})

	DescribeTableSubtree("cli stack image",
		func(entry cliStackEntry) {
			var image string

			It("image present in snapshot", func() {
				image = snapshotData.Images[entry.snapshotKey]
				Expect(image).NotTo(BeEmpty(), "image key %q not found in snapshot", entry.snapshotKey)
			})

			for _, bin := range entry.binaries {
				It(fmt.Sprintf("contains valid %s binary for %s/%s", filepath.Base(bin.pathInImage), bin.osName, bin.arch), func() {
					Expect(image).NotTo(BeEmpty(), "image not loaded — prerequisite 'image present in snapshot' failed")

					workDir := filepath.Join(tmpDir, entry.snapshotKey, bin.osName, bin.arch, filepath.Base(bin.pathInImage))
					Expect(os.MkdirAll(workDir, 0755)).To(Succeed())

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()

					archivePath := filepath.Join(workDir, filepath.Base(bin.pathInImage))
					Expect(support.FileFromImage(ctx, image, bin.pathInImage, workDir)).To(
						Succeed(), "extracting %s from %s", bin.pathInImage, entry.snapshotKey,
					)

					// All archives (including .exe.gz) are tar archives — extract the first file.
					executablePath, err := support.ExtractFirstFileFromTarGz(archivePath, workDir)
					Expect(err).NotTo(HaveOccurred(), "extracting archive %s", archivePath)

					Expect(executablePath).NotTo(BeEmpty())
					Expect(verifyBinaryExecutable(executablePath, bin.osName, bin.arch)).To(Succeed())
				})
			}
		},
		Entry("cosign-cli-stack-image", cliStackEntries[0]),
		Entry("gitsign-cli-stack-image", cliStackEntries[1]),
		Entry("rekor-cli-stack-image", cliStackEntries[2]),
		Entry("fetch-tsa-certs-cli-stack-image", cliStackEntries[3]),
		Entry("trillian-cli-stack-image", cliStackEntries[4]),
		Entry("tuftool-cli-stack-image", cliStackEntries[5]),
		Entry("conforma-cli-stack-image", cliStackEntries[6]),
		Entry("model-transparency-cli-stack-image", cliStackEntries[7]),
	)
})
