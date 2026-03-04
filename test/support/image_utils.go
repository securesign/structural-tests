package support

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type ImageData struct {
	Image  string
	Labels map[string]string
}

func PullImageIfNotPresentLocally(ctx context.Context, imageDefinition string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create docker client: %w", err)
	}
	defer cli.Close()

	// Inspect the image to check if it exists locally.
	_, _, err = cli.ImageInspectWithRaw(ctx, imageDefinition)
	if err == nil {
		return nil // image is present already
	}

	if client.IsErrNotFound(err) {
		log.Printf("Image '%s' not found locally, pulling...\n", imageDefinition)
		pullResp, pullErr := cli.ImagePull(ctx, imageDefinition, image.PullOptions{
			Platform: "linux/amd64",
		})
		if pullErr != nil {
			return fmt.Errorf("failed to pull image: %w", pullErr)
		}
		defer pullResp.Close()
		// ensure the pull operation completes.
		if _, err := io.Copy(io.Discard, pullResp); err != nil {
			return fmt.Errorf("failed to read pull response: %w", err)
		}
		return nil
	}

	// another type of error, return it.
	return fmt.Errorf("failed to inspect image: %w", err)
}

func RunImage(imageDefinition string, entrypoint, commands []string) (string, error) {
	ctx := context.TODO()
	err := PullImageIfNotPresentLocally(ctx, imageDefinition)
	if err != nil {
		return "", err
	}

	log.Printf("Running image %s with commands %v\n", imageDefinition, commands)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("error while initializing docker client: %w", err)
	}

	config := &container.Config{
		Image: imageDefinition,
		Cmd:   commands,
	}
	if len(entrypoint) > 0 {
		config.Entrypoint = entrypoint
	}

	resp, err := cli.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed while creating container: %w", err)
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed while starting container: %w", err)
	}

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, "")
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("failed while waiting for container to finish: %w", err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("cannot get container logs: %w", err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, out)
	if err != nil {
		return "", fmt.Errorf("getting logs from the stream failed: %w", err)
	}

	return buf.String(), nil
}

func InspectImageForLabels(imageDefinition string) (map[string]string, error) {
	ctx := context.TODO()
	err := PullImageIfNotPresentLocally(ctx, imageDefinition)
	if err != nil {
		return nil, err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("error while initializing docker client: %w", err)
	}
	defer cli.Close()

	inspectData, _, err := cli.ImageInspectWithRaw(ctx, imageDefinition)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect image %s: %w", imageDefinition, err)
	}
	if inspectData.Config == nil || len(inspectData.Config.Labels) == 0 {
		log.Printf("Image [%s] does not have any labels\n", imageDefinition)
		return make(map[string]string), nil
	}

	return inspectData.Config.Labels, nil
}

func GetImageLabel(imageDefinition, labelName string) (string, error) {
	labels, err := InspectImageForLabels(imageDefinition)
	if err != nil {
		return "", err
	}
	if labelValue, ok := labels[labelName]; ok {
		return labelValue, nil
	}
	return "", fmt.Errorf("label [%s] not found in image %s", labelName, imageDefinition)
}

func FileFromImage(ctx context.Context, imageName, filePath, outputPath string) error {
	err := PullImageIfNotPresentLocally(ctx, imageName)
	if err != nil {
		return err
	}

	// Initialize the Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create a container from the image
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"echo", "dummy"},
	}, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Ensure container removal
	defer func() {
		_ = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{
			Force: true,
		})
	}()

	// Use Docker's API to copy the file from the container's filesystem
	reader, _, err := cli.CopyFromContainer(ctx, resp.ID, filePath)
	if err != nil {
		return fmt.Errorf("failed to copy file from container: %w", err)
	}
	defer reader.Close()

	outputFilePath := filepath.Join(outputPath, filepath.Base(filePath))
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	return extractFileFromTar(reader, outFile)
}

func extractFileFromTar(reader io.Reader, writer io.Writer) error {
	// Tar reader to read the file from the tar stream
	tarReader := tar.NewReader(reader)

	// Extract the file to the output path
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Check if it's the file we want
		if header.Typeflag == tar.TypeReg {
			// Copy the content of the file
			_, err = io.Copy(writer, tarReader) //nolint:gosec
			if err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
			return nil
		}
	}
	return errors.New("failed to find file in tar")
}

// GetAnsibleCollectionArchiveFromImage copies /releases from the image, finds
// redhat-artifact_signer*.tar.gz in the tar stream, and returns its content.
func GetAnsibleCollectionArchiveFromImage(ctx context.Context, imageName string) ([]byte, error) {
	err := PullImageIfNotPresentLocally(ctx, imageName)
	if err != nil {
		return nil, err
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"echo", "dummy"},
	}, nil, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}
	defer func() {
		_ = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}()
	reader, _, err := cli.CopyFromContainer(ctx, resp.ID, AnsibleCollectionPathInImage)
	if err != nil {
		return nil, fmt.Errorf("copy %s from container: %w", AnsibleCollectionPathInImage, err)
	}
	defer reader.Close()
	return findAndReadCollectionArchiveFromTar(reader)
}

func findAndReadCollectionArchiveFromTar(reader io.Reader) ([]byte, error) {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(header.Name)
		if strings.HasPrefix(base, "redhat-artifact_signer") && strings.HasSuffix(base, ".tar.gz") {
			b, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", header.Name, err)
			}
			log.Printf("Found collection archive in image: %s (%d bytes)\n", header.Name, len(b))
			return b, nil
		}
	}
	return nil, errors.New("redhat-artifact_signer*.tar.gz not found in image /releases")
}

// manifestListPlatform is the platform field in a manifest list entry.
type manifestListPlatform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
}

// manifestListEntry is one entry in the manifests array.
type manifestListEntry struct {
	Digest   string               `json:"digest"`
	Platform manifestListPlatform `json:"platform"`
}

// manifestListOutput is the JSON output of podman/docker manifest inspect.
type manifestListOutput struct {
	Manifests []manifestListEntry `json:"manifests"`
}

// ResolveManifestListForPlatform returns the image ref (repo@digest) for the given platform
// by inspecting the manifest list. imageRef is the manifest list ref (e.g. quay.io/...@sha256:...).
// platform is e.g. "linux/amd64" or "linux/arm64". Returns the same ref if the image is not
// a manifest list (e.g. single-arch) or if resolution fails.
const platformPartCount = 2 // os/arch

func ResolveManifestListForPlatform(ctx context.Context, imageRef, platform string) (string, error) {
	parts := strings.SplitN(platform, "/", platformPartCount)
	if len(parts) != platformPartCount {
		return imageRef, fmt.Errorf("platform must be os/arch, got %q", platform)
	}
	wantOS, wantArch := parts[0], parts[1]

	// Try podman first, then docker
	var out []byte
	var err error
	for _, cmdName := range []string{"podman", "docker"} {
		cmd := exec.CommandContext(ctx, cmdName, "manifest", "inspect", imageRef)
		out, err = cmd.Output()
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return imageRef, fmt.Errorf("manifest inspect failed (tried podman and docker): %w", err)
	}

	var list manifestListOutput
	if err := json.Unmarshal(out, &list); err != nil {
		return imageRef, fmt.Errorf("parse manifest list: %w", err)
	}
	if len(list.Manifests) == 0 {
		return imageRef, errors.New("manifest list has no manifests")
	}

	for _, entry := range list.Manifests {
		if entry.Platform.OS == wantOS && entry.Platform.Architecture == wantArch {
			repo := imageRef
			if at := strings.Index(imageRef, "@"); at != -1 {
				repo = imageRef[:at]
			}
			return repo + "@" + entry.Digest, nil
		}
	}
	return imageRef, fmt.Errorf("no manifest for platform %s", platform)
}
