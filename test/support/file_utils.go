package support

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	testroot "github.com/securesign/structural-tests/test"
)

func GetFileContent(filePath string) ([]byte, error) {
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

func downloadFileContent(url string, accessToken string) ([]byte, error) {
	log.Printf("Downloading file %s\n", url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	if accessToken != "" {
		req.Header.Add("Authorization", "token "+accessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("error status: %s", resp.Status)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return body, nil
}

func loadFileContent(filePath string) ([]byte, error) {
	log.Printf("Loading file %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	contentBuffer, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read content of the file: %w", err)
	}
	return contentBuffer, nil
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

// LoadAnsibleCollectionFromImage extracts the collection archive from the image
// at /releases/redhat-artifact_signer-*.tar.gz and returns the content of the
// given file inside that archive (e.g. roles/tas_single_node/defaults/main.yml).
func LoadAnsibleCollectionFromImage(ctx context.Context, imageRef, ansibleImagesFile string) ([]byte, error) {
	archiveBytes, err := GetAnsibleCollectionArchiveFromImage(ctx, imageRef)
	if err != nil {
		return nil, err
	}
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("gunzip collection archive: %w", err)
	}
	defer gzReader.Close()
	return lookThroughTarFile(gzReader, ansibleImagesFile)
}

func lookThroughTarFile(reader io.Reader, filePath string) ([]byte, error) {
	tarReader := tar.NewReader(reader)
	for {
		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read files from tar file: %w", err)
		}

		if tarHeader.Name == filePath {
			log.Printf("Found %s\n", tarHeader.Name)
			tarFileContent, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read from tar file: %w", err)
			}
			if tarFileContent == nil {
				log.Printf("%s not found\n", filePath)
			}
			return tarFileContent, nil
		}
	}
	return nil, nil
}
