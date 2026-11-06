package envguard

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const maxFileSize = 1024 * 1024

var defaultSkipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
}

func checkGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git is not installed or not found in PATH")
	}
	return nil
}

func isGitRepo(root string) bool {
	if err := checkGit(); err != nil {
		return false
	}
	cmd := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

func gitCmd(root string, args ...string) (string, error) {
	if err := checkGit(); err != nil {
		return "", err
	}
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return "", fmt.Errorf("git %s: %s", args[0], bytes.TrimSpace(exit.Stderr))
	}
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func isBinaryBytes(b []byte) bool {
	limit := 8192
	if len(b) < limit {
		limit = len(b)
	}
	return bytes.IndexByte(b[:limit], 0) != -1
}

func isBinaryOrLarge(fullPath string) (bool, error) {
	fi, err := os.Stat(fullPath)
	if err != nil {
		return true, err
	}
	if fi.IsDir() {
		return true, nil
	}
	if fi.Size() > maxFileSize {
		return true, nil
	}
	f, err := os.Open(fullPath)
	if err != nil {
		return true, err
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, nil
	}
	return isBinaryBytes(buf[:n]), nil
}

func shouldSkipDir(name string) bool {
	return defaultSkipDirs[name]
}

func shouldSkipPath(relSlash string) bool {
	parts := strings.Split(relSlash, "/")
	for _, p := range parts {
		if defaultSkipDirs[p] {
			return true
		}
	}
	return false
}

func loadNosubignore(root string) []string {
	p := filepath.Join(root, ".nosubignore")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var pats []string
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		pats = append(pats, trimmed)
	}
	return pats
}

func isIgnored(patterns []string, relSlash string) bool {
	for _, pat := range patterns {
		if matchPattern(pat, relSlash) {
			return true
		}
	}
	return false
}

func matchPattern(pat, relSlash string) bool {
	pat = strings.TrimSpace(pat)
	if pat == "" || strings.HasPrefix(pat, "#") {
		return false
	}
	pat = strings.TrimPrefix(pat, "./")
	pat = strings.TrimSuffix(pat, "/")
	if m, _ := path.Match(pat, relSlash); m {
		return true
	}
	if m, _ := path.Match(pat+"/*", relSlash); m {
		return true
	}
	parts := strings.Split(relSlash, "/")
	for i := 1; i <= len(parts); i++ {
		sub := strings.Join(parts[:i], "/")
		if m, _ := path.Match(pat, sub); m {
			return true
		}
	}
	return false
}

func parseHunkLine(hunk string) int {
	idx := strings.Index(hunk, "+")
	if idx == -1 {
		return 1
	}
	rest := hunk[idx+1:]
	var numStr strings.Builder
	for _, r := range rest {
		if r >= '0' && r <= '9' {
			numStr.WriteRune(r)
		} else {
			break
		}
	}
	if n, err := strconv.Atoi(numStr.String()); err == nil {
		return n
	}
	return 1
}
