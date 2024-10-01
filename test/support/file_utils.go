package support

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	testroot "github.com/securesign/structural-tests/test"
)

func GetFileContent(filePath string) (string, error) {
	snapshotFile, isLocal := checkFilePath(filePath)
	if isLocal {
		return loadFileContent(snapshotFile)
	}
	githubToken := GetEnvAsSecret(EnvTestGithubToken)
	return downloadFileContent(snapshotFile, githubToken)
}

func checkFilePath(filePath string) (string, bool) {
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return filePath, false
	}
	filePath = localPathCleanup(filePath)
	return filePath, true
}

func localPathCleanup(origPath string) string {
	finalPath := origPath
	if !filepath.IsAbs(origPath) {
		finalPath = filepath.Join(testroot.GetRootPath(), origPath)
	}
	return filepath.Clean(finalPath)
}

func downloadFileContent(url string, accessToken string) (string, error) {
	log.Printf("Downloading file %s\n", url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}
	if accessToken != "" {
		req.Header.Add("Authorization", "token "+accessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := errors.New("bad status: " + resp.Status)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(body), nil
}

func loadFileContent(filePath string) (string, error) {
	log.Printf("Loading file %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	contentBuffer, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read content of the file: %w", err)
	}
	return string(contentBuffer), nil
}

// DecompressGzipFile decompresses a Gzip file and writes the decompressed content to a specified output file.
func DecompressGzipFile(gzipPath string, outputPath string) error {
	// Open the Gzip file
	gzipFile, err := os.Open(gzipPath)
	if err != nil {
		return fmt.Errorf("failed to open Gzip file: %w", err)
	}
	defer gzipFile.Close()

	// Create a Gzip reader
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return fmt.Errorf("failed to create Gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create the output file to write the decompressed data
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Copy the decompressed data from the Gzip reader to the output file
	_, err = io.Copy(outputFile, gzipReader) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to write decompressed data: %w", err)
	}
	return nil
}
