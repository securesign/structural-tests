package support

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func RunImage(imageDefinition string, commands []string) (string, error) {
	ctx := context.TODO()
	log.Printf("Running image %s with commands %v\n", imageDefinition, commands)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("error while initializing docker client: %w", err)
	}
	reader, err := cli.ImagePull(ctx, imageDefinition, image.PullOptions{
		Platform: "linux/amd64",
	})
	if err != nil {
		return "", fmt.Errorf("cannot pull image %s: %w", imageDefinition, err)
	}
	_, _ = io.Copy(os.Stdout, reader)
	_ = reader.Close()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageDefinition,
		Cmd:   commands,
	}, nil, nil, nil, "")
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

func FileFromImage(ctx context.Context, imageName, filePath, outputPath string) error {
	// Initialize the Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("cannot pull image %s: %w", imageName, err)
	}
	_, _ = io.Copy(io.Discard, reader)
	_ = reader.Close()

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
	reader, _, err = cli.CopyFromContainer(ctx, resp.ID, filePath)
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
