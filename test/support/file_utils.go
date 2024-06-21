package support

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	testroot "github.com/securesign/structural-tests/test"
)

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
		finalPath = filepath.Join(testroot.RootPath, origPath)
	}
	return filepath.Clean(finalPath)
}

func downloadFileContent(url string, accessToken string) (string, error) {
	log.Printf("Downloading file %s\n", url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
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
