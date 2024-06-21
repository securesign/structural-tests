package testroot

import (
	"path/filepath"
	"runtime"
)

var (
	_, file, _, _ = runtime.Caller(0)
	RootPath      = filepath.Join(filepath.Dir(file), "..")
)
