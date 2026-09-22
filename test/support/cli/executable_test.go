package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVerifyBinaryExecutable(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyBinaryExecutable(executable, runtime.GOOS, runtime.GOARCH); err != nil {
		t.Fatal(err)
	}
	if err := verifyBinaryExecutable(executable, runtime.GOOS, "unknown"); err == nil {
		t.Fatal("expected architecture mismatch")
	}
	if err := verifyBinaryExecutable(executable, "unknown", runtime.GOARCH); err == nil {
		t.Fatal("expected unsupported OS error")
	}
	invalid := filepath.Join(t.TempDir(), "invalid")
	if err := os.WriteFile(invalid, []byte("not an executable"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, osName := range []string{"linux", "darwin", "windows"} {
		if err := verifyBinaryExecutable(invalid, osName, "amd64"); err == nil {
			t.Fatalf("expected invalid %s executable to fail", osName)
		}
	}
}
