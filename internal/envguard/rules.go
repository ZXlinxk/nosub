package envguard

import (
	"math"
	"regexp"
	"strings"
)

type patternRule struct {
	name   string
	re     *regexp.Regexp
	filter func(match string) bool
}

var (
	specificRules = []patternRule{
		{
			name: "aws-access-key",
			re:   regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			filter: func(m string) bool {
				return !strings.Contains(m, "EXAMPLE")
			},
		},
		{
			name: "github-token",
			re:   regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}|github_pat_[a-zA-Z0-9_]{22,}`),
		},
		{
			name: "gitlab-token",
			re:   regexp.MustCompile(`glpat-[a-zA-Z0-9_-]{20,}`),
		},
		{
			name: "slack-token",
			re:   regexp.MustCompile(`xox[abeoprs]-[a-zA-Z0-9_-]{10,}`),
		},
		{
			name: "stripe-key",
			re:   regexp.MustCompile(`[sr]k_live_[0-9a-zA-Z]{16,}`),
		},
		{
			name: "anthropic-key",
			re:   regexp.MustCompile(`sk-ant-[a-zA-Z0-9_-]{32,}`),
		},
		{
			name: "openai-key",
			re:   regexp.MustCompile(`sk-(?:proj-)?[A-Za-z0-9_-]{32,}`),
			filter: func(m string) bool {
				return !strings.HasPrefix(m, "sk-ant-")
			},
		},
		{
			name: "google-api-key",
			re:   regexp.MustCompile(`AIza[a-zA-Z0-9_-]{35}`),
		},
		{
			name: "sendgrid-key",
			re:   regexp.MustCompile(`SG\.[a-zA-Z0-9_-]{16,}\.[a-zA-Z0-9_-]{16,}`),
		},
		{
			name: "npm-token",
			re:   regexp.MustCompile(`npm_[A-Za-z0-9]{36}`),
		},
		{
			name: "pypi-token",
			re:   regexp.MustCompile(`pypi-[a-zA-Z0-9_-]{50,}`),
		},
		{
			name: "private-key",
			re:   regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
		},
		{
			name: "jwt",
			re:   regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}`),
		},
	}

	urlPasswordRe   = regexp.MustCompile(`[a-z][a-z0-9+.-]*://[^\s:@/]+:([^\s:@/]+)@`)
	awsSecretValRe  = regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`)
	dottedNameRe    = regexp.MustCompile(`^[A-Za-z_]+(\.[A-Za-z_]+)+$`)
	secretKeyWordRe = regexp.MustCompile(`(?i)(?:^|[^a-z0-9])(secret|token|password|passwd|pwd|api_key|apikey|private_key|auth_key)(?:$|[^a-z0-9])`)
)

type hit struct{ rule, secret string }

func scanLine(path, line string, lineNum int) []Finding {
	if strings.Contains(line, "nosub:allow") {
		return nil
	}
	hits := matchLine(line)
	if len(hits) == 0 {
		return nil
	}
	secrets := make([]string, len(hits))
	for i, h := range hits {
		secrets[i] = h.secret
	}
	// Every secret is masked in every preview, a line can hold more than one.
	preview := makePreview(line, secrets)
	findings := make([]Finding, len(hits))
	for i, h := range hits {
		findings[i] = Finding{File: path, Line: lineNum, Rule: h.rule, Preview: preview}
	}
	return findings
}

func matchLine(line string) []hit {
	var hits []hit
	for _, r := range specificRules {
		for _, m := range r.re.FindAllString(line, -1) {
			if r.filter == nil || r.filter(m) {
				hits = append(hits, hit{r.name, m})
			}
		}
	}

	key, val, ok := parseAssignment(line)
	if ok && !strings.Contains(val, "EXAMPLE") && awsSecretValRe.MatchString(val) {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "aws") && strings.Contains(lower, "secret") {
			hits = append(hits, hit{"aws-secret-key", val})
		}
	}

	for _, m := range urlPasswordRe.FindAllStringSubmatch(line, -1) {
		if !isPlaceholder(m[1]) {
			hits = append(hits, hit{"url-password", m[1]})
		}
	}

	if len(hits) == 0 && ok && isSecretKeyName(key) && isLiteral(val) &&
		!isPlaceholder(val) && len(val) >= 12 && shannonEntropy(val) >= 3.0 {
		hits = append(hits, hit{"generic-secret", val})
	}
	return hits
}

// isLiteral rejects code on the right side of an assignment, like getPassword(u) or user.password.
func isLiteral(val string) bool {
	return isASCII(val) && !strings.ContainsAny(val, "()[]{} ") && !dottedNameRe.MatchString(val)
}

// isASCII filters out translated strings, secrets are not written in Japanese.
func isASCII(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r > 127 }) < 0
}

func parseAssignment(line string) (key, val string, ok bool) {
	idx := -1
	for i := 0; i < len(line) && idx < 0; i++ {
		switch {
		case line[i] == '=':
			idx = i
		case line[i] == ':' && !strings.HasPrefix(line[i:], "://"):
			idx = i
		}
	}
	if idx < 0 {
		return "", "", false
	}

	left := line[:idx]
	right := line[idx+1:]

	left = strings.TrimSpace(left)
	left = strings.TrimPrefix(left, "export ")
	left = strings.TrimPrefix(left, "const ")
	left = strings.TrimPrefix(left, "let ")
	left = strings.TrimPrefix(left, "var ")
	left = strings.Trim(left, `"'`)
	left = strings.TrimSpace(left)
	// A typed declaration such as "String password" keeps only the name.
	if f := strings.Fields(left); len(f) > 1 {
		left = f[len(f)-1]
	}

	val = unwrapVal(right)
	return left, val, true
}

func unwrapVal(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, ";,")
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') ||
			(s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func isSecretKeyName(key string) bool {
	return secretKeyWordRe.MatchString(key)
}

func isPlaceholder(val string) bool {
	val = strings.TrimSpace(val)
	if val == "" {
		return true
	}
	val = unwrapVal(val)
	if val == "" {
		return true
	}
	lower := strings.ToLower(val)
	switch lower {
	case "changeme", "password", "passwd", "pass", "pw", "pwd", "secret", "example", "sample", "test", "dummy":
		return true
	}
	if strings.HasPrefix(lower, "your_") || strings.HasPrefix(lower, "your-") {
		return true
	}
	if isRepeatedChar(val) {
		return true
	}
	if strings.HasPrefix(val, "<") && strings.HasSuffix(val, ">") {
		return true
	}
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		return true
	}
	if strings.HasPrefix(val, "$") || strings.HasPrefix(val, "{") || strings.HasPrefix(val, "%") {
		return true
	}
	if strings.Contains(lower, "os.getenv") ||
		strings.Contains(lower, "process.env") ||
		strings.Contains(lower, "getenv") ||
		strings.Contains(lower, "env(") {
		return true
	}
	return false
}

func isRepeatedChar(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}

func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}
	n := float64(len([]rune(s)))
	var entropy float64
	for _, count := range counts {
		p := float64(count) / n
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func makePreview(line string, secrets []string) string {
	preview := strings.TrimSpace(line)
	for _, s := range secrets {
		r := []rune(s)
		preview = strings.ReplaceAll(preview, s, string(r[:min(4, len(r))])+"****")
	}
	if len(preview) > 120 {
		preview = strings.ToValidUTF8(preview[:120], "")
	}
	return preview
}
