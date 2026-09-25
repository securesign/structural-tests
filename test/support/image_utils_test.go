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
		{name: "native arch", pulledArch: "amd64", platform: "linux/amd64"},
		{name: "arm64 arch", pulledArch: "arm64", platform: "linux/arm64"},
		{name: "foreign arch", pulledArch: "s390x", platform: "linux/s390x"},
		{name: "uncached foreign arch", pulledArch: "ppc64le", platform: "linux/ppc64le"},
		{name: "wrong pulled arch", pulledArch: "amd64", platform: "linux/arm64", wantErr: "expected linux/arm64"},
		{name: "missing platform", platform: "linux/s390x", pullFailure: true, wantErr: "no matching manifest"},
		{name: "missing file", pulledArch: "amd64", platform: "linux/amd64", missingFile: true, wantErr: "file missing"},
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
			if !fixture.pulled {
				t.Fatal("platform-specific image was not pulled")
			}
		})
	}
}

type imageExtractionCase struct {
	name, pulledArch, platform, wantErr string
	pullFailure, missingFile            bool
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
		fixture.inspectImage(writer, path)
	case path == "/images/create":
		fixture.pulled = true
		if got := request.URL.Query().Get("platform"); got != fixture.testCase.platform {
			fixture.t.Errorf("pull platform=%q, want %q", got, fixture.testCase.platform)
		}
		if fixture.testCase.pullFailure {
			_, _ = writer.Write([]byte(`{"errorDetail":{"message":"no matching manifest"},"error":"no matching manifest"}`))
			return
		}
		_, _ = writer.Write([]byte(`{"status":"done"}`))
	case path == "/containers/create":
		fixture.createContainer(writer, request)
	case path == "/containers/test-container/json":
		fixture.inspectContainer(writer)
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

func (fixture *imageExtractionFixture) inspectImage(writer http.ResponseWriter, path string) {
	if !strings.HasPrefix(path, "/images/verified-") {
		fixture.t.Error("inspected multi-platform reference instead of selected image")
	}
	arch := strings.TrimSuffix(strings.TrimPrefix(path, "/images/verified-"), "/json")
	_, _ = fmt.Fprintf(writer, `{"Id":"verified-%s","Os":"linux","Architecture":%q}`, arch, arch)
}

func (fixture *imageExtractionFixture) inspectContainer(writer http.ResponseWriter) {
	_, _ = fmt.Fprintf(writer, `{"Image":"verified-%s"}`, fixture.testCase.pulledArch)
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
	var body container.Config
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		fixture.t.Error(err)
	}
	if body.Image != "example/image:release" {
		fixture.t.Errorf("container image=%q, want original reference", body.Image)
	}
	if got := request.URL.Query().Get("platform"); got != fixture.testCase.platform {
		fixture.t.Errorf("create platform=%q, want %q", got, fixture.testCase.platform)
	}
	fixture.created = true
	_, _ = writer.Write([]byte(`{"Id":"test-container"}`))
}

func TestPlatformManifestReference(t *testing.T) {
	indexRef := "example/image@sha256:" + strings.Repeat("a", 64)
	childDigest := "sha256:" + strings.Repeat("b", 64)
	index := fmt.Sprintf(`{"manifests":[
		{"digest":"sha256:%s","platform":{"os":"linux","architecture":"amd64"}},
		{"digest":%q,"platform":{"os":"linux","architecture":"arm64"}}
	]}`, strings.Repeat("c", 64), childDigest)
	for _, testCase := range []struct {
		name, manifest, arch, wantRef string
		wantErr                       bool
	}{
		{name: "exact platform", manifest: index, arch: "arm64", wantRef: "example/image@" + childDigest},
		{name: "missing platform", manifest: index, arch: "s390x", wantErr: true},
		{name: "empty index", manifest: `{"manifests":[]}`, arch: "arm64", wantErr: true},
		{name: "invalid JSON", manifest: `{`, arch: "arm64", wantErr: true},
		{
			name: "single image", manifest: `{"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`,
			arch: "arm64", wantRef: indexRef,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := platformManifestReference([]byte(testCase.manifest), indexRef, "linux", testCase.arch)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("error=%v, want error=%t", err, testCase.wantErr)
			}
			if !testCase.wantErr && got != testCase.wantRef {
				t.Fatalf("reference=%q, want %q", got, testCase.wantRef)
			}
		})
	}
}
