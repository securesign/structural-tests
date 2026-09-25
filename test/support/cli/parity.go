package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/securesign/structural-tests/test/support"
)

const runtimeDirectoryMode = 0700

func compareRuntimeBinary(ctx context.Context, images map[string]string, stackKey string, binary StackBinary, executable, workDir string) error {
	runtime := binary.Runtime
	image := images[runtime.ImageKey]
	if image == "" {
		return fmt.Errorf("archive %s:%s: runtime image key %q not found in snapshot", stackKey, binary.Path, runtime.ImageKey)
	}
	platform := binary.OS + "/" + binary.Arch
	description := fmt.Sprintf("%s (%s):%s versus %s (%s):%s for %s",
		stackKey, images[stackKey], binary.Path, runtime.ImageKey, image, runtime.Path, platform)
	// Pull the platform manifest by its own digest so other tests keep the snapshot index reference.
	runtimeImage, err := support.ResolveManifestListForPlatform(ctx, image, platform)
	if err != nil {
		return fmt.Errorf("%s: resolve runtime image: %w", description, err)
	}
	runtimeDir := filepath.Join(workDir, "runtime")
	if err := os.MkdirAll(runtimeDir, runtimeDirectoryMode); err != nil {
		return fmt.Errorf("%s: %w", description, err)
	}
	if err := support.FileFromImageForPlatform(ctx, runtimeImage, runtime.Path, runtimeDir, platform); err != nil {
		return fmt.Errorf("%s: extract runtime binary: %w", description, err)
	}
	if err := compareBinaryFiles(executable, filepath.Join(runtimeDir, filepath.Base(runtime.Path))); err != nil {
		return fmt.Errorf("%s: %w", description, err)
	}
	return nil
}

func compareBinaryFiles(archivePath, runtimePath string) error {
	stackHash, err := support.ChecksumFile(archivePath)
	if err != nil {
		return fmt.Errorf("checksum: %w", err)
	}
	runtimeHash, err := support.ChecksumFile(runtimePath)
	if err != nil {
		return fmt.Errorf("checksum: %w", err)
	}
	if !bytes.Equal(stackHash, runtimeHash) {
		return fmt.Errorf("binary mismatch: archive SHA-256=%x, runtime SHA-256=%x", stackHash, runtimeHash)
	}
	return nil
}
