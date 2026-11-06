//go:build !windows

package i18n

import (
	"os/exec"
	"runtime"
	"strings"
)

func systemLocale() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	out, err := exec.Command("defaults", "read", "-g", "AppleLocale").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
