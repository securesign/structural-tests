package support

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

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
	reader, err := cli.ImagePull(ctx, imageDefinition, image.PullOptions{})
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
