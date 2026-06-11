package platform

import (
	"fmt"
	"runtime"
)

// Current returns the platform key used by the registry (e.g. "linux-amd64").
func Current() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// Validate checks if a platform string looks reasonable.
func Validate(p string) bool {
	return p != ""
}
