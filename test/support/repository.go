package support

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
)

const (
	groupRegistry = iota + 1
	groupImage
	groupTag
)

var containerRegexp = regexp.MustCompile(`^(?P<registry>[\w.\-_]+)/(?P<image>[^@:]+)(?P<tag>.+)$`)

type Repository struct {
	Name      string `json:"repository"`
	ID        string `json:"_id"` //nolint:tagliatelle
	Published bool   `json:"published"`
}

type RepositoryList struct {
	Data []Repository `json:"data"`
}

func (r *RepositoryList) FindByImage(image string) *Repository {
	match := containerRegexp.FindStringSubmatch(image)
	if len(match) != groupTag+1 {
		return nil
	}

	for _, rep := range r.Data {
		if rep.Name == match[groupImage] {
			return &rep
		}
	}
	return nil
}

func LoadRepositoryList() (*RepositoryList, error) {
	fileName := GetEnv(EnvRepositoriesFile)
	if fileName == "" {
		fileName = DefaultRepositoriesFile
		log.Printf("using default repositories file %s\n", DefaultRepositoriesFile)
	}
	content, err := GetFileContent(fileName)
	if err != nil {
		return nil, err
	}
	var list RepositoryList
	err = json.Unmarshal(content, &list)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repositories file: %w", err)
	}
	return &list, nil
}
