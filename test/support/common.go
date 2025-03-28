package support

import (
	"fmt"
	"log"
	"os"
)

func GetEnv(key string) string {
	return getEnv(key, false)
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

func GetMapKeys(m map[string]string) []string {
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

func LogMap(message string, data map[string]string) {
	result := message + "\n"
	for key, value := range data {
		result += fmt.Sprintf("    [%-41s] %s\n", key, value)
	}
	log.Print(result)
}
