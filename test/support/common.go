package support

import (
	"fmt"
	"log"
	"os"
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

func GetMapValues(m map[string]string) []string {
	var result []string
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
