package support

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"gopkg.in/yaml.v3"
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

func ValidateJson(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var js interface{}
	return json.Unmarshal(content, &js)
}

func ValidateYaml(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var yml interface{}
	return yaml.Unmarshal(content, &yml)
}
