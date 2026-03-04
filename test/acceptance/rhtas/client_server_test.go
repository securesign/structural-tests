package acceptance

import (
	"context"
	"crypto/sha256"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
)

const (
	cliServerFileMask = "%s-%s.gz"
	cliServerPathMask = "/var/www/html/clients/%s/" + cliServerFileMask

	cliImageBasePath = "/usr/local/bin"
	cliImageFileMask = "%s_cli_%s_%s.gz"

	osLinux   = "linux"
	osDarwin  = "darwin"
	osWindows = "windows"

	cliCosign         = "cosign"
	cliGitsign        = "gitsign"
	cliRekorCli       = "rekor-cli"
	cliFetchTsaCerts  = "fetch-tsa-certs"
	cliCreatetree     = "createtree"
	cliUpdatetree     = "updatetree"
	cliTuftool        = "tuftool"
)

// Multiarch CLIs: built per-arch (manifest list in snapshot). From Dockerfile.clients.rh.
// Single-image CLIs (ec, tuftool) are not in this list.
var multiArchCLISnapshotKeys = []string{ //nolint:gochecknoglobals // test CLI snapshot keys
	"cosign-cli-image",
	"gitsign-cli-image",
	"fetch-tsa-certs-cli-image",
	"rekor-cli-image",
	"createtree-image",
	"updatetree-image",
}

func snapshotKeyForCLI(cli string) string {
	switch cli {
	case cliCreatetree, cliUpdatetree:
		return cli + "-image"
	case cliTuftool:
		return "tuf-tool-image"
	case cliRekorCli:
		return "rekor-cli-image"
	default:
		return cli + "-cli-image"
	}
}

func isMultiArchImageKey(key string) bool {
	for _, k := range multiArchCLISnapshotKeys {
		if k == key {
			return true
		}
	}
	return false
}

// sourcePathInImageMultiArch returns the path inside the CLI source image for the given (os, arch).
// Only for multiarch CLIs (cosign, gitsign, rekor-cli, fetch-tsa-certs, createtree, updatetree).
//
//nolint:funlen // path lookup table per CLI/OS
func sourcePathInImageMultiArch(cli, osName, arch string) string {
	base := cliImageBasePath + "/"
	switch cli {
	case cliCosign:
		switch osName {
		case osLinux:
			return base + "cosign.gz"
		case osDarwin:
			return base + "cosign-darwin-" + arch + ".gz"
		case osWindows:
			return base + "cosign-windows-amd64.exe.gz"
		}
	case cliGitsign:
		switch osName {
		case osLinux:
			return base + "gitsign_cli_linux.gz"
		case osDarwin:
			return base + "gitsign_cli_darwin_" + arch + ".gz"
		case osWindows:
			return base + "gitsign_cli_windows_amd64.exe.gz"
		}
	case cliRekorCli:
		switch osName {
		case osLinux:
			return base + "rekor_cli_linux.gz"
		case osDarwin:
			return base + "rekor_cli_darwin_" + arch + ".gz"
		case osWindows:
			return base + "rekor_cli_windows_amd64.exe.gz"
		}
	case cliFetchTsaCerts:
		switch osName {
		case osLinux:
			return base + "fetch_tsa_certs_linux.gz"
		case osDarwin:
			return base + "fetch_tsa_certs_darwin_" + arch + ".gz"
		case osWindows:
			return base + "fetch_tsa_certs_windows_amd64.exe.gz"
		}
	case cliCreatetree:
		switch osName {
		case osLinux:
			return base + "createtree.gz"
		case osDarwin:
			return base + "createtree-darwin-" + arch + ".gz"
		case osWindows:
			return base + "createtree-windows-amd64.exe.gz"
		}
	case cliUpdatetree:
		switch osName {
		case osLinux:
			return base + "updatetree.gz"
		case osDarwin:
			return base + "updatetree-darwin-" + arch + ".gz"
		case osWindows:
			return base + "updatetree-windows-amd64.exe.gz"
		}
	}
	return ""
}

// multiArchCLIs is the list of CLI names that use multiarch (manifest list) source images.
var multiArchCLIs = []string{cliCosign, cliGitsign, cliRekorCli, cliFetchTsaCerts, cliCreatetree, cliUpdatetree} //nolint:gochecknoglobals // test CLI list

var _ = Describe("Client server", Ordered, func() {

	var clientServerImage string
	var snapshotData support.SnapshotData
	var tmpDir string
	var serverChecksums map[string][]byte // key: "cli/osName/arch", populated by verify Its (all CLIs)

	Describe("client-server image", func() {
		It("snapshot.json", func() {
			var err error
			snapshotData, err = support.ParseSnapshotData()
			Expect(err).NotTo(HaveOccurred())

			clientServerImage = snapshotData.Images["client-server-image"]
			Expect(clientServerImage).NotTo(BeEmpty())
		})

		It("", func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "client-server")
			Expect(err).NotTo(HaveOccurred())
			serverChecksums = make(map[string][]byte)
			log.Printf("TEMP directory: %s", tmpDir)
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
					var image string
					var gzipServerSHA []byte

					It("init", func() {
						switch cli {
						case cliCreatetree, cliUpdatetree:
							image = snapshotData.Images[cli+"-image"]
						case cliTuftool:
							image = snapshotData.Images["tuf-tool-image"]
						case cliRekorCli:
							image = snapshotData.Images["rekor-cli-image"]
						default:
							image = snapshotData.Images[cli+"-cli-image"]
						}
					})

					It(fmt.Sprintf("verify %s-%s executable", osName, arch), func() {
						osPath := filepath.Join(tmpDir, osName)
						Expect(os.MkdirAll(osPath, 0755)).To(Succeed())

						By("get gzip file from container image")
						ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
						defer cancel()
						Expect(support.FileFromImage(ctx, clientServerImage, fmt.Sprintf(cliServerPathMask, osName, cli, arch), osPath)).To(Succeed())

						By("decompress gzip")
						gzipPath := filepath.Join(osPath, fmt.Sprintf(cliServerFileMask, cli, arch))
						targetPath := filepath.Join(osPath, cli+"-"+arch)
						Expect(support.DecompressGzipFile(gzipPath, targetPath)).To(Succeed())

						By("verify executable OS and arch")
						executable := filepath.Join(tmpDir, osName, cli+"-"+arch)
						Expect(verifyExecutable(executable, osName, arch)).To(Succeed())

						By("checksums of gzip file")
						var err error
						gzipServerSHA, err = checksumFile(filepath.Join(osPath, fmt.Sprintf(cliServerFileMask, cli, arch)))
						Expect(err).NotTo(HaveOccurred())
						serverChecksums[cli+"/"+osName+"/"+arch] = append([]byte(nil), gzipServerSHA...)
					})

					if support.IsBeforeVersion("1.4.0") || !isMultiArchImageKey(snapshotKeyForCLI(cli)) {
						It(fmt.Sprintf("compare checkum of %s-%s with source image", osName, arch), func() {
							var (
								targetPath = tmpDir
								fileName   string
								filePath   string
							)

							switch cli {
							case cliTuftool:
								Skip("`tuftool` do not have gzip in source image")
							case "ec":
								Skip("`ec` source image is not part of handover")
							case cliCosign, cliUpdatetree, cliCreatetree:
								if osName == osWindows {
									fileName = fmt.Sprintf("%s-%s-%s.exe.gz", cli, osName, arch)
								} else {
									fileName = fmt.Sprintf("%s-%s-%s.gz", cli, osName, arch)
								}
							case cliRekorCli, cliFetchTsaCerts:
								ncli := strings.ReplaceAll(cli, "-", "_")
								if osName == osWindows {
									fileName = fmt.Sprintf("%s_%s_%s.exe.gz", ncli, osName, arch)
								} else {
									fileName = fmt.Sprintf("%s_%s_%s.gz", ncli, osName, arch)
								}
							default:
								if osName == osWindows {
									fileName = fmt.Sprintf(cliImageFileMask, cli, osName, arch+".exe")
								} else {
									fileName = fmt.Sprintf(cliImageFileMask, cli, osName, arch)
								}
							}
							filePath = filepath.Join(cliImageBasePath, fileName)

							By("get gzip file from container image")
							ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
							defer cancel()

							Expect(support.FileFromImage(ctx, image, filePath, targetPath)).To(Succeed())

							By("checksums of gzip file")
							gzipImageSHA, err := checksumFile(filepath.Join(targetPath, fileName))
							Expect(err).NotTo(HaveOccurred())

							By("compare checksum with client server file")
							Expect(gzipImageSHA).To(Equal(gzipServerSHA))
						})
					}
				}
			}
		},
		Entry("cosign", cliCosign, support.GetOSArchMatrix()),
		Entry("gitsign", cliGitsign, support.GetOSArchMatrix()),
		Entry("rekor-cli", cliRekorCli, support.GetOSArchMatrix()),
		Entry("ec", "ec", support.GetOSArchMatrix()),
		Entry("fetch-tsa-certs", cliFetchTsaCerts, support.GetOSArchMatrix()),
		Entry("tuftool", cliTuftool, map[string][]string{
			osLinux: {"amd64"},
		}),
		Entry("updatetree", cliUpdatetree, support.GetOSArchMatrix()),
		Entry("createtree", cliCreatetree, support.GetOSArchMatrix()),
	)

	It("compare all multiarch binaries (all CLIs) with source images", func() {
		if support.IsBeforeVersion("1.4.0") {
			Skip("multiarch comparison only for version 1.4.0 and later")
		}
		matrix := support.GetOSArchMatrix()
		var errMsgs []string
		for _, cli := range multiArchCLIs {
			if !isMultiArchImageKey(snapshotKeyForCLI(cli)) {
				continue
			}
			for osName, archs := range matrix {
				for _, arch := range archs {
					sourceKey := snapshotKeyForCLI(cli)
					sourceImage := snapshotData.Images[sourceKey]
					clientServerPath := fmt.Sprintf(cliServerPathMask, osName, cli, arch)
					cliImagePath := sourcePathInImageMultiArch(cli, osName, arch)
					platform := "linux/" + arch

					resolvedImage, err := support.ResolveManifestListForPlatform(context.Background(), sourceImage, platform)
					if err != nil {
						errMsgs = append(errMsgs, fmt.Sprintf("%s %s-%s: resolve manifest: %v", cli, osName, arch, err))
						continue
					}

					log.Printf("%s %s", clientServerImage, clientServerPath)
					log.Printf("%s %s", resolvedImage, cliImagePath)
					log.Printf("from %s", sourceImage)

					srcDir := filepath.Join(tmpDir, "multiarch-src", cli, osName, arch)
					if err := os.MkdirAll(srcDir, 0755); err != nil {
						errMsgs = append(errMsgs, fmt.Sprintf("%s %s-%s: mkdir: %v", cli, osName, arch, err))
						continue
					}
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
					if err := support.FileFromImage(ctx, resolvedImage, cliImagePath, srcDir); err != nil {
						cancel()
						errMsgs = append(errMsgs, fmt.Sprintf("%s %s-%s: copy from image: %v", cli, osName, arch, err))
						continue
					}
					cancel()

					sourceGzipPath := filepath.Join(srcDir, filepath.Base(cliImagePath))
					gzipSourceSHA, err := checksumFile(sourceGzipPath)
					if err != nil {
						errMsgs = append(errMsgs, fmt.Sprintf("%s %s-%s: checksum: %v", cli, osName, arch, err))
						continue
					}
					key := cli + "/" + osName + "/" + arch
					gzipServerSHA, ok := serverChecksums[key]
					if !ok {
						errMsgs = append(errMsgs, fmt.Sprintf("%s %s-%s: no server checksum", cli, osName, arch))
						continue
					}
					if string(gzipSourceSHA) != string(gzipServerSHA) {
						errMsgs = append(errMsgs, fmt.Sprintf("%s %s-%s: gzip checksum mismatch\n  client-server SHA256: %x\n  source image SHA256: %x",
							cli, osName, arch, gzipServerSHA, gzipSourceSHA))
					}
				}
			}
		}
		if len(errMsgs) > 0 {
			Fail("multiarch checksum failures:\n\n" + strings.Join(errMsgs, "\n\n"))
		}
	})
})

// verifyExecutable verifies that the executable file matches the target OS and architecture.
func verifyExecutable(filePath, osName, arch string) error {
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

// checksumFile computes the SHA256 checksum of a given file.
func checksumFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to hash file %s: %w", filePath, err)
	}

	return hasher.Sum(nil), nil
}
