package support

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
)

func GetEnv(key string) string {
	return getEnv(key, false)
}

// VersionForConfig returns VERSION env if set; otherwise tries to infer version from SNAPSHOT path
// (e.g. "../releases-1.3.2/1.3.2/stable/snapshot.json" -> "1.3.2"). Used so 1.2.x and 1.3.x config overrides apply without setting VERSION.
func VersionForConfig() string {
	if v := GetEnv(EnvVersion); v != "" {
		return v
	}
	snapshotPath := GetEnv(EnvReleasesSnapshotFile)
	if snapshotPath == "" {
		return ""
	}
	// Match semver in path (e.g. 1.2.2, 1.3.2); take last match in case path has multiple numbers.
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := re.FindAllString(snapshotPath, -1)
	if len(matches) > 0 {
		return matches[len(matches)-1]
	}
	return ""
}

func GetEnvAsSecret(key string) string {
	return getEnv(key, true)
}

func getEnv(key string, isSecret bool) string {
	envValue, _ := os.LookupEnv(key)
	var logMessage string
	if isSecret && envValue != "" {
		logMessage = fmt.Sprintf("%s=%s", key, "*****")
	} else {
		logMessage = fmt.Sprintf("%s='%s'", key, envValue)
	}
	log.Println(logMessage)
	return envValue
}

func GetMapKeys[V any](m map[string]V) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func GetMapValues(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func GetMapKeysSorted[V any](m map[string]V) []string {
	keys := GetMapKeys(m)
	slices.Sort(keys)
	return keys
}

func SplitMap(original map[string]string, keysToKeep []string) (map[string]string, map[string]string) {
	remaining := make(map[string]string)
	moved := make(map[string]string)

	for key, value := range original {
		if contains(keysToKeep, key) {
			remaining[key] = value
		} else {
			moved[key] = value
		}
	}
	return remaining, moved
}

func contains(source []string, value string) bool {
	for _, v := range source {
		if v == value {
			return true
		}
	}
	return false
}

func LogArray(message string, data []string) {
	result := message + "\n"
	for _, value := range data {
		result += fmt.Sprintf("    %s\n", value)
	}
	log.Print(result)
}

func LogMap[V any](message string, data map[string]V) {
	result := message + "\n"
	for key, value := range data {
		result += fmt.Sprintf("    [%-41v] %v\n", key, value)
	}
	log.Print(result)
}

func LogMapByProvidedKeys[V any](message string, data map[string]V, keysToLog []string) {
	result := message + "\n"
	for _, key := range keysToLog {
		result += fmt.Sprintf("    [%-53v] %v\n", key, data[key])
	}
	log.Print(result)
}
