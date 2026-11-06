package svgjsx

import (
	"path/filepath"
	"strings"
	"unicode"
)

// ComponentName derives a PascalCase React component name from a file path.
func ComponentName(path string) string {
	base := filepath.Base(path)
	if strings.HasSuffix(strings.ToLower(base), ".svg") {
		base = base[:len(base)-4]
	}

	var words []string
	var cur strings.Builder
	for _, r := range base {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}

	if len(words) == 0 {
		return "Icon"
	}

	var res strings.Builder
	if unicode.IsDigit(rune(words[0][0])) {
		res.WriteString("Icon")
	}

	for _, w := range words {
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		res.WriteString(string(runes))
	}

	return res.String()
}
