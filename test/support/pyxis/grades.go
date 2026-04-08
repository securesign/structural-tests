package pyxis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const baseURL = "https://catalog.redhat.com/api/containers/v1"

// Grade represents a Red Hat container image freshness grade (A through F).
type Grade byte

const (
	GradeA Grade = 'A'
	GradeB Grade = 'B'
	GradeC Grade = 'C'
	GradeD Grade = 'D'
	GradeE Grade = 'E'
	GradeF Grade = 'F'
)

var containerRegex = regexp.MustCompile(`^(?P<registry>[\w.\-_]+)/(?P<image>[^@:]+)(?P<tag>.+)$`)

func ParseGrade(s string) (Grade, bool) {
	if len(s) == 1 && s[0] >= 'A' && s[0] <= 'F' {
		return Grade(s[0]), true
	}
	return 0, false
}

func (g Grade) String() string {
	return string(g)
}

// WorseThan returns true when g is strictly worse than other.
func (g Grade) WorseThan(other Grade) bool {
	return g > other
}

const (
	groupRegistry = iota + 1
	groupImage
	groupTag
)

// FreshnessGrade represents a single grade period from the Pyxis API.
type FreshnessGrade struct {
	Grade        string  `json:"grade"`
	CreationDate string  `json:"creation_date"` //nolint:tagliatelle // Pyxis API uses snake_case
	StartDate    string  `json:"start_date"`    //nolint:tagliatelle // Pyxis API uses snake_case
	EndDate      *string `json:"end_date"`      //nolint:tagliatelle // Pyxis API uses snake_case
}

// ImageGradeInfo holds per-architecture grade data returned by the Pyxis images endpoint.
type ImageGradeInfo struct {
	ID              string           `json:"_id"` //nolint:tagliatelle // Pyxis API uses _id
	Architecture    string           `json:"architecture"`
	FreshnessGrades []FreshnessGrade `json:"freshness_grades"` //nolint:tagliatelle // Pyxis API uses snake_case
}

type imagesResponse struct {
	Data []ImageGradeInfo `json:"data"`
}

// CurrentGrade returns the currently active grade based on the freshness_grades schedule.
func (img ImageGradeInfo) CurrentGrade() string {
	now := time.Now().UTC()
	for _, freshness := range img.FreshnessGrades {
		start, err := parseTime(freshness.StartDate)
		if err != nil || start.After(now) {
			continue
		}
		if freshness.EndDate == nil {
			return freshness.Grade
		}
		end, err := parseTime(*freshness.EndDate)
		if err != nil {
			continue
		}
		if now.Before(end) {
			return freshness.Grade
		}
	}
	return ""
}

func parseTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse pyxis time value: %s", value)
	}
	return t.UTC(), nil
}

// FetchImageGrades queries Pyxis for grade information about a specific image
// identified by its manifest digest. Returns per-architecture grade data.
// Tries manifest_list_digest first, then falls back to manifest_schema2_digest.
func FetchImageGrades(digest string) ([]ImageGradeInfo, error) {
	grades, err := fetchByDigestFilter("repositories.manifest_list_digest", digest)
	if err != nil {
		return nil, err
	}
	if len(grades) > 0 {
		return grades, nil
	}
	return fetchByDigestFilter("repositories.manifest_schema2_digest", digest)
}

func fetchByDigestFilter(field, digest string) ([]ImageGradeInfo, error) {
	apiURL := fmt.Sprintf("%s/images?filter=%s==%s&include=data._id,data.architecture,data.freshness_grades&page_size=10",
		baseURL, field, digest)
	log.Printf("Fetching image grades from %s\n", apiURL)

	const requestTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch image grades: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pyxis API returned %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result imagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result.Data, nil
}

func extractDigest(imageRef string) (string, error) {
	match := containerRegex.FindStringSubmatch(imageRef)
	if len(match) != groupTag+1 {
		return "", fmt.Errorf("cannot parse image reference: %s", imageRef)
	}
	tag := match[groupTag]
	if strings.HasPrefix(tag, "@") {
		return tag[1:], nil
	}
	return "", fmt.Errorf("image reference does not contain a digest: %s", imageRef)
}

func extractRegistryAndRepository(imageRef string) (string, string) {
	match := containerRegex.FindStringSubmatch(imageRef)
	if len(match) != groupTag+1 {
		return "", ""
	}
	return match[groupRegistry], match[groupImage]
}

// GradeResults holds the outcome of a bulk grade lookup.
type GradeResults struct {
	// Grades maps "registry/repository" to per-architecture grade data.
	Grades map[string][]ImageGradeInfo
	// NotFound lists "registry/repository" entries that Pyxis had no data for.
	NotFound []string
}

// FetchGradesForImages fetches Pyxis grade data for each unique image digest found in the
// image map. Images not found in Pyxis are recorded in GradeResults.NotFound.
func FetchGradesForImages(images map[string]string) (*GradeResults, error) {
	res := &GradeResults{Grades: make(map[string][]ImageGradeInfo)}
	seen := make(map[string]bool)
	for key, imageRef := range images {
		digest, err := extractDigest(imageRef)
		if err != nil {
			return nil, fmt.Errorf("failed to extract digest for %s (%s): %w", key, imageRef, err)
		}
		if seen[digest] {
			continue
		}
		seen[digest] = true

		grades, err := FetchImageGrades(digest)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch grades for %s (%s): %w", key, imageRef, err)
		}

		registry, repository := extractRegistryAndRepository(imageRef)
		repoKey := registry + "/" + repository
		if len(grades) == 0 {
			log.Printf("Not found in Pyxis: %s (digest %s)\n", repoKey, digest)
			res.NotFound = append(res.NotFound, repoKey)
			continue
		}

		res.Grades[repoKey] = grades
		log.Printf("Fetched grades for %s (%s): %d arch entries\n", repoKey, digest, len(grades))
	}
	return res, nil
}

// ValidateGrades checks that no image will have a grade worse than threshold
// at any point between now and now+days.
func ValidateGrades(gradeResults map[string][]ImageGradeInfo, threshold Grade, days int) []error {
	var errs []error
	deadline := time.Now().UTC().AddDate(0, 0, days)
	now := time.Now().UTC()

	for repo, images := range gradeResults {
		for _, img := range images {
			for _, freshness := range img.FreshnessGrades {
				if err := checkFreshness(freshness, threshold, now, deadline, repo, img.Architecture); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	return errs
}

func checkFreshness(freshness FreshnessGrade, threshold Grade, now, deadline time.Time, repo, arch string) error {
	grade, ok := ParseGrade(freshness.Grade)
	if ok && !grade.WorseThan(threshold) {
		return nil
	}
	startDate, err := parseTime(freshness.StartDate)
	if err != nil {
		return nil //nolint:nilerr // unparseable dates are skipped, not treated as failures
	}
	if startDate.After(deadline) {
		return nil
	}
	if freshness.EndDate != nil {
		endDate, err := parseTime(*freshness.EndDate)
		if err != nil {
			return nil //nolint:nilerr // unparseable dates are skipped, not treated as failures
		}
		if !endDate.After(now) {
			return nil
		}
	}
	if !startDate.After(now) {
		return fmt.Errorf("%s (arch=%s): current grade is %s", repo, arch, freshness.Grade)
	}
	return fmt.Errorf("%s (arch=%s): grade will drop to %s on %s", repo, arch, freshness.Grade, freshness.StartDate)
}
