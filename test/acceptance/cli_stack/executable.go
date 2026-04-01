package acceptance

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
)

func verifyBinaryExecutable(filePath, osName, arch string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	switch osName {
	case "linux":
		elfFile, err := elf.NewFile(file)
		if err != nil {
			return fmt.Errorf("failed to read ELF file %s: %w", filePath, err)
		}
		if elfFile.FileHeader.Machine != elfArchType(arch) {
			return fmt.Errorf("architecture mismatch: expected %s, got %s", arch, elfFile.FileHeader.Machine.String())
		}
	case "windows":
		peFile, err := pe.NewFile(file)
		if err != nil {
			return fmt.Errorf("failed to read PE file %s: %w", filePath, err)
		}
		if peFile.FileHeader.Machine != peMachineType(arch) {
			return fmt.Errorf("architecture mismatch: expected %s, got %v", arch, peFile.FileHeader.Machine)
		}
	case "darwin":
		machoFile, err := macho.NewFile(file)
		if err != nil {
			return fmt.Errorf("failed to read Mach-O file %s: %w", filePath, err)
		}
		if machoFile.Cpu != machoCpuType(arch) {
			return fmt.Errorf("architecture mismatch: expected %s, got %v", arch, machoFile.Cpu)
		}
	default:
		return fmt.Errorf("unsupported OS: %s", osName)
	}
	return nil
}

func elfArchType(arch string) elf.Machine {
	switch arch {
	case "amd64":
		return elf.EM_X86_64
	case "arm64":
		return elf.EM_AARCH64
	case "ppc64le":
		return elf.EM_PPC64
	case "s390x":
		return elf.EM_S390
	default:
		return elf.EM_NONE
	}
}

func peMachineType(arch string) uint16 {
	switch arch {
	case "amd64":
		return pe.IMAGE_FILE_MACHINE_AMD64
	case "arm64":
		return pe.IMAGE_FILE_MACHINE_ARM64
	default:
		return pe.IMAGE_FILE_MACHINE_UNKNOWN
	}
}

func machoCpuType(arch string) macho.Cpu {
	switch arch {
	case "amd64":
		return macho.CpuAmd64
	case "arm64":
		return macho.CpuArm64
	default:
		return 0
	}
}
