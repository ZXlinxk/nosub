package envguard

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// IsEnvFile reports whether the given file path represents an environment file.
func IsEnvFile(path string) bool {
	base := filepath.Base(filepath.ToSlash(path))
	lower := strings.ToLower(base)

	nonEnvSuffixes := []string{".example", ".sample", ".template", ".dist", ".bak", ".backup"}
	for _, s := range nonEnvSuffixes {
		if strings.HasSuffix(lower, s) {
			return false
		}
	}

	if lower == ".env" || strings.HasPrefix(lower, ".env.") || strings.HasSuffix(lower, ".env") {
		return true
	}
	return false
}

// GenerateExample generates a sanitized .env.example template from env content.
func GenerateExample(content string) string {
	var sb strings.Builder
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			break
		}
		crlf := strings.HasSuffix(line, "\r")
		cleanLine := strings.TrimSuffix(line, "\r")

		trimmed := strings.TrimSpace(cleanLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			sb.WriteString(cleanLine)
		} else {
			eqIdx := strings.Index(cleanLine, "=")
			if eqIdx == -1 {
				sb.WriteString(cleanLine)
			} else {
				left := cleanLine[:eqIdx]
				val := cleanLine[eqIdx+1:]

				keyName := strings.TrimSpace(left)
				keyName = strings.TrimPrefix(keyName, "export ")
				keyName = strings.TrimSpace(keyName)

				trimmedVal := strings.TrimSpace(val)
				if !isSensitiveKey(keyName) && isPreservedVal(trimmedVal) {
					sb.WriteString(cleanLine)
				} else {
					sb.WriteString(left)
					sb.WriteString("=")
				}
			}
		}

		if crlf {
			sb.WriteString("\r")
		}
		sb.WriteString("\n")
	}

	res := sb.String()
	if !strings.HasSuffix(content, "\n") && strings.HasSuffix(res, "\n") {
		res = strings.TrimSuffix(res, "\n")
		res = strings.TrimSuffix(res, "\r")
	}
	return res
}

func isPreservedVal(val string) bool {
	lower := strings.ToLower(unwrapVal(val))
	switch lower {
	case "true", "false", "yes", "no", "on", "off":
		return true
	}
	if _, err := strconv.ParseFloat(lower, 64); err == nil {
		return true
	}
	return false
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	words := []string{"secret", "token", "password", "passwd", "pwd", "key", "auth", "cred", "private", "hash", "cert"}
	for _, w := range words {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// EnsureGitignore adds .env patterns to .gitignore if not present.
func EnsureGitignore(root string) (bool, error) {
	p := filepath.Join(root, ".gitignore")
	b, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	targets := []string{".env", ".env.*", "!.env.example"}
	existingLines := make(map[string]bool)
	var content string

	if err == nil {
		content = string(b)
		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			existingLines[strings.TrimSpace(scanner.Text())] = true
		}
	}

	var missing []string
	for _, t := range targets {
		if !existingLines[t] {
			missing = append(missing, t)
		}
	}

	if len(missing) == 0 {
		return false, nil
	}

	var sb strings.Builder
	sb.WriteString(content)
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	for _, m := range missing {
		sb.WriteString(m)
		sb.WriteString("\n")
	}

	if err := os.WriteFile(p, []byte(sb.String()), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
