package acceptance

import (
	"context"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

const cliServerPathMask = "/var/www/html/clients/%s/%s-%s.gz"

var _ = Describe("Client server", Ordered, func() {

	var clientServerImage string
	var tmpDir string

	Describe("client-server image", func() {
		It("snapshot.json", func() {
			snapshotImages, err := support.ParseSnapshotImages()
			Expect(err).NotTo(HaveOccurred())

			clientServerImage = snapshotImages["client-server-image"]
			Expect(clientServerImage).NotTo(BeEmpty())
		})

		It("", func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "client-server")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	AfterAll(func() {
		DeferCleanup(func() {
			err := os.RemoveAll(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	DescribeTableSubtree("cli",
		func(cli string, matrix support.OSArchMatrix) {
			for osName, archs := range matrix {
				for _, arch := range archs {
					It(fmt.Sprintf("verify %s-%s executable", osName, arch), func() {
						osPath := filepath.Join(tmpDir, osName)
						Expect(os.MkdirAll(osPath, 0755)).To(Succeed())

						By("get gzip file from container image")
						ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
						defer cancel()
						Expect(support.FileFromImage(ctx, clientServerImage, fmt.Sprintf(cliServerPathMask, osName, cli, arch), osPath)).To(Succeed())

						By("decompress gzip")
						gzipPath := filepath.Join(osPath, cli+"-"+arch+".gz")
						targetPath := filepath.Join(osPath, cli+"-"+arch)
						Expect(support.DecompressGzipFile(gzipPath, targetPath)).To(Succeed())

						By("verify executable OS and arch")
						executable := filepath.Join(tmpDir, osName, cli+"-"+arch)
						Expect(verifyExecutable(executable, osName, arch)).To(Succeed())
					})
				}
			}

		},
		Entry("cosign", "cosign", support.GetOSArchMatrix()),
		Entry("gitsign", "gitsign", support.GetOSArchMatrix()),
		Entry("rekor-cli", "rekor-cli", support.GetOSArchMatrix()),
		Entry("ec", "ec", support.GetOSArchMatrix()),
		Entry("fetch-tsa-certs", "fetch-tsa-certs", support.GetOSArchMatrix()),
		Entry("tuftool", "tuftool", map[string][]string{
			"linux": {"amd64"},
		}),
		Entry("updatetree", "updatetree", support.GetOSArchMatrix()),
		Entry("createtree", "createtree", support.GetOSArchMatrix()),
	)
})

// verifyExecutable verifies that the executable file matches the target OS and architecture.
func verifyExecutable(filePath, osName, arch string) error { //nolint:cyclop
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	switch osName {
	case "linux":
		// Use debug/elf to inspect ELF files (Linux)
		elfFile, err := elf.NewFile(file)
		if err != nil {
			return fmt.Errorf("failed to read ELF file %s: %w", filePath, err)
		}
		if elfFile.FileHeader.Machine != getELFArchType(arch) {
			return fmt.Errorf("architecture mismatch: expected %s, got %s", arch, elfFile.FileHeader.Machine.String())
		}
	case "windows":
		// Use debug/pe to inspect PE files (Windows)
		peFile, err := pe.NewFile(file)
		if err != nil {
			return fmt.Errorf("failed to read PE file %s: %w", filePath, err)
		}
		if peFile.FileHeader.Machine != getPEMachineType(arch) {
			return fmt.Errorf("architecture mismatch: expected %s, got %v", arch, peFile.FileHeader.Machine)
		}
	case "darwin":
		// Use debug/macho to inspect Mach-O files (macOS)
		machoFile, err := macho.NewFile(file)
		if err != nil {
			return fmt.Errorf("failed to read Mach-O file %s: %w", filePath, err)
		}
		if machoFile.Cpu != getMachOCpuType(arch) {
			return fmt.Errorf("architecture mismatch: expected %s, got %v", arch, machoFile.Cpu)
		}
	default:
		return fmt.Errorf("unsupported OS: %s", osName)
	}
	return nil
}

// getELFArchType returns the appropriate ELF arch type for the given architecture.
func getELFArchType(arch string) elf.Machine {
	switch arch {
	case "x86_64", "amd64": //nolint:goconst
		return elf.EM_X86_64
	case "arm64": //nolint:goconst
		return elf.EM_AARCH64
	case "ppc64le":
		return elf.EM_PPC64
	case "s390x":
		return elf.EM_S390
	default:
		return elf.EM_NONE
	}
}

// getPEMachineType returns the appropriate PE machine type for the given architecture.
func getPEMachineType(arch string) uint16 {
	switch arch {
	case "x86_64", "amd64":
		return pe.IMAGE_FILE_MACHINE_AMD64
	case "arm64":
		return pe.IMAGE_FILE_MACHINE_ARM64
	default:
		return pe.IMAGE_FILE_MACHINE_UNKNOWN
	}
}

// getMachOCpuType returns the appropriate Mach-O CPU type for the given architecture.
func getMachOCpuType(arch string) macho.Cpu {
	switch arch {
	case "x86_64", "amd64":
		return macho.CpuAmd64
	case "arm64":
		return macho.CpuArm64
	default:
		return 0 // Unsupported architecture
	}
}
