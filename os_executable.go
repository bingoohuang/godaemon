//go:build go1.8
// +build go1.8

package godaemon

import (
	"os"
)

func osExecutable() (string, error) {
	return os.Executable()
}
