package support

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

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

func GetFileContent(filePath string) (string, error) {
	snapshotFile, isLocal := checkFilePath(filePath)
	if isLocal {
		return loadFileContent(snapshotFile)
	} else {
		githubToken := GetEnvOrDefaultSecret(EnvTestGithubToken, "")
		return downloadFileContent(snapshotFile, githubToken)
	}
}

func checkFilePath(filePath string) (string, bool) {
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return filePath, false
	} else {
		filePath = localPathCleanup(filePath)
		return filePath, true
	}
}

func localPathCleanup(origPath string) string {
	finalPath := origPath
	if !filepath.IsAbs(origPath) {
		// not ideal solution
		// want to have path relative to the project directory
		// without test/acceptance-tests
		finalPath = filepath.Join("..", "..", origPath)
	}
	return filepath.Clean(finalPath)
}

func downloadFileContent(url string, accessToken string) (string, error) {
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

func loadFileContent(filePath string) (string, error) {
	log.Printf("Loading file %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	contentBuffer, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(contentBuffer), nil
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

func GetMapValues(m map[string]string) []string {
	var result []string
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
