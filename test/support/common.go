package support

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"io"
	"log"
	"net/http"
	"os"
)

func GetEnv(key string) string {
	return getEnvOrDefault(key, "", true)
}

func GetEnvOrDefault(key, defaultValue string) string {
	return getEnvOrDefault(key, defaultValue, true)
}

func GetEnvOrDefaultSecret(key, defaultValue string) string {
	return getEnvOrDefault(key, defaultValue, false)
}

func getEnvOrDefault(key, defaultValue string, isLogged bool) string {
	var returnValue string
	isDefaultValue := false
	value, exists := os.LookupEnv(key)
	if !exists && defaultValue != "" {
		returnValue = defaultValue
		isDefaultValue = true
	} else {
		returnValue = value
	}
	var logMessage string
	if isLogged || returnValue == "" {
		logMessage = fmt.Sprintf("%s='%s'", key, returnValue)
	} else {
		logMessage = fmt.Sprintf("%s=%s", key, "*****")
	}
	if isDefaultValue {
		logMessage = fmt.Sprintf("%s (default)", logMessage)
	}
	log.Println(logMessage)
	return returnValue
}

func DownloadFileContent(url string, accessToken string) (string, error) {
	log.Printf("Downloading file %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if accessToken != "" {
		req.Header.Add("Authorization", "token "+accessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func GitCloneWithAuth(url string, branch string, auth transport.AuthMethod) (string, *git.Repository, error) {
	dir, err := os.MkdirTemp("", "securesign-")
	if err != nil {
		return "", nil, err
	}
	log.Println(fmt.Sprintf("Cloning %s on branch %s to %s", url, branch, dir))
	cloneOptions := &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	}
	if auth != nil {
		cloneOptions.Auth = auth
	}
	repo, err := git.PlainClone(dir, false, cloneOptions)
	if err == nil {
		log.Println("Cloned successfully")
	}
	return dir, repo, err
}

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
