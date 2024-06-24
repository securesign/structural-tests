package testroot

import (
	"log"
	"path/filepath"
	"runtime"
)

func GetRootPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Printf("Caller file may not be correctly recovered: %s", file)
	}
	return filepath.Join(filepath.Dir(file), "..")
}
