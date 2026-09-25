package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareBinaryFiles(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "archive")
	runtime := filepath.Join(dir, "runtime")
	for _, testCase := range []struct{ name, content, wantErr string }{
		{"identical", "same binary", ""},
		{"different", "different binary", "binary mismatch: archive SHA-256="},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := os.WriteFile(archive, []byte("same binary"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(runtime, []byte(testCase.content), 0600); err != nil {
				t.Fatal(err)
			}
			err := compareBinaryFiles(archive, runtime)
			if testCase.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.wantErr) || !strings.Contains(err.Error(), "runtime SHA-256=") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
	for _, paths := range [][2]string{{"missing", runtime}, {archive, "missing"}, {dir, runtime}, {archive, dir}} {
		if err := compareBinaryFiles(paths[0], paths[1]); err == nil {
			t.Fatalf("expected read failure for %v", paths)
		}
	}
}

func TestRuntimeFailures(t *testing.T) {
	binary := StackBinary{Path: "/binaries/tool.tar.gz", OS: "linux", Arch: "arm64", Runtime: &RuntimeBinary{ImageKey: "runtime-image", Path: "/tool"}}
	images := map[string]string{"stack-image": "stack@sha256:abc"}
	err := compareRuntimeBinary(context.Background(), images, "stack-image", binary, "unused", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), `runtime image key "runtime-image" not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Setenv("PATH", "")
	images["runtime-image"] = "runtime@sha256:def"
	err = compareRuntimeBinary(context.Background(), images, "stack-image", binary, "unused", t.TempDir())
	for _, want := range []string{"resolve runtime image", "stack@sha256:abc", "runtime@sha256:def", "linux/arm64", "/tool"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("error %v missing %q", err, want)
		}
	}
}
