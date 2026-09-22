package support

import (
	"archive/tar"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestFileFromImageForPlatform(t *testing.T) {
	for _, testCase := range []imageExtractionCase{
		{name: "cached single arch", cachedArch: "amd64", platform: "linux/amd64"},
		{name: "cached foreign arch", cachedArch: "s390x", platform: "linux/s390x"},
		{name: "replace wrong cache", cachedArch: "amd64", pulledArch: "arm64", platform: "linux/arm64"},
		{name: "uncached foreign arch", pulledArch: "ppc64le", platform: "linux/ppc64le"},
		{name: "wrong pulled arch", cachedArch: "amd64", pulledArch: "amd64", platform: "linux/arm64", wantErr: "expected linux/arm64"},
		{name: "missing platform", platform: "linux/s390x", pullFailure: true, wantErr: "no matching manifest"},
		{name: "missing file", cachedArch: "amd64", platform: "linux/amd64", missingFile: true, wantErr: "file missing"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := &imageExtractionFixture{t: t, testCase: testCase}
			server := httptest.NewServer(fixture)
			defer server.Close()
			t.Setenv("DOCKER_HOST", server.URL)
			t.Setenv("DOCKER_API_VERSION", "1.45")
			dir := t.TempDir()
			err := FileFromImageForPlatform(context.Background(), "example/image:release", "/tool", dir, testCase.platform)
			if testCase.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.wantErr) {
					t.Fatalf("error=%v, want %q", err, testCase.wantErr)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(filepath.Join(dir, "tool"))
				if err != nil || string(data) != "binary" {
					t.Fatalf("extracted %q: %v", data, err)
				}
			}
			if fixture.created != fixture.removed {
				t.Fatal("container was not cleaned up")
			}
			if want := testCase.cachedArch != strings.TrimPrefix(testCase.platform, "linux/"); fixture.pulled != want {
				t.Fatalf("pulled=%v, want %v", fixture.pulled, want)
			}
		})
	}
}

type imageExtractionCase struct {
	name, cachedArch, pulledArch, platform, wantErr string
	pullFailure, missingFile                        bool
}

type imageExtractionFixture struct {
	t                        *testing.T
	testCase                 imageExtractionCase
	pulled, created, removed bool
}

func (fixture *imageExtractionFixture) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, "/v1.45")
	writer.Header().Set("Content-Type", "application/json")
	switch {
	case strings.HasPrefix(path, "/images/") && strings.HasSuffix(path, "/json"):
		arch := fixture.testCase.cachedArch
		if fixture.pulled {
			arch = fixture.testCase.pulledArch
		}
		if arch == "" {
			http.Error(writer, `{"message":"image missing"}`, http.StatusNotFound)
			return
		}
		_, _ = fmt.Fprintf(writer, `{"Id":"verified-%s","Os":"linux","Architecture":%q}`, arch, arch)
	case path == "/images/create":
		fixture.pulled = true
		if got := request.URL.Query().Get("platform"); got != fixture.testCase.platform {
			fixture.t.Errorf("pull platform=%q, want %q", got, fixture.testCase.platform)
		}
		if fixture.testCase.pullFailure {
			http.Error(writer, `{"message":"no matching manifest"}`, http.StatusNotFound)
			return
		}
		_, _ = writer.Write([]byte(`{"status":"done"}`))
	case path == "/containers/create":
		fixture.createContainer(writer, request)
	case path == "/containers/test-container/archive":
		if request.URL.Query().Get("path") != "/tool" {
			fixture.t.Error("wrong executable path")
		}
		if fixture.testCase.missingFile {
			http.Error(writer, `{"message":"file missing"}`, http.StatusNotFound)
			return
		}
		fixture.writeArchive(writer)
	case path == "/containers/test-container" && request.Method == http.MethodDelete:
		fixture.removed = true
		writer.WriteHeader(http.StatusNoContent)
	default:
		fixture.t.Errorf("unexpected Docker request: %s %s", request.Method, path)
		writer.WriteHeader(http.StatusBadRequest)
	}
}

func (fixture *imageExtractionFixture) writeArchive(writer http.ResponseWriter) {
	writer.Header().Set("X-Docker-Container-Path-Stat", base64.StdEncoding.EncodeToString([]byte(`{"name":"tool","size":6,"mode":493}`)))
	writer.Header().Set("Content-Type", "application/x-tar")
	tarWriter := tar.NewWriter(writer)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "tool", Mode: 0755, Size: 6}); err != nil {
		fixture.t.Error(err)
	}
	if _, err := tarWriter.Write([]byte("binary")); err != nil {
		fixture.t.Error(err)
	}
	if err := tarWriter.Close(); err != nil {
		fixture.t.Error(err)
	}
}

func (fixture *imageExtractionFixture) createContainer(writer http.ResponseWriter, request *http.Request) {
	fixture.created = true
	var body container.Config
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		fixture.t.Error(err)
	}
	arch := strings.TrimPrefix(fixture.testCase.platform, "linux/")
	if body.Image != "verified-"+arch {
		fixture.t.Errorf("container image=%q, expected verified ID", body.Image)
	}
	if got := request.URL.Query().Get("platform"); got != fixture.testCase.platform {
		fixture.t.Errorf("create platform=%q, want %q", got, fixture.testCase.platform)
	}
	_, _ = writer.Write([]byte(`{"Id":"test-container"}`))
}
