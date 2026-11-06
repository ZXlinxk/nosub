package svgjsx

import (
	"fmt"
	"strings"
	"unicode"
)

type styleDecl struct {
	key string
	val string
}

func parseStyle(s string) []styleDecl {
	var decls []styleDecl
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.IndexByte(part, ':')
		if idx == -1 {
			continue
		}
		k := strings.TrimSpace(part[:idx])
		v := strings.TrimSpace(part[idx+1:])
		if k != "" {
			decls = append(decls, styleDecl{key: k, val: v})
		}
	}
	return decls
}

func formatStyleObject(decls []styleDecl) string {
	if len(decls) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("style={{ ")
	for i, d := range decls {
		if i > 0 {
			b.WriteString(", ")
		}
		k := d.key
		if strings.HasPrefix(k, "--") {
			fmt.Fprintf(&b, "%q", k)
		} else {
			b.WriteString(toCamelCase(k))
		}
		fmt.Fprintf(&b, ": %q", d.val)
	}
	b.WriteString(" }}")
	return b.String()
}

func toCamelCase(s string) string {
	var b strings.Builder
	upper := false
	for i, r := range s {
		if r == '-' || r == ':' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else if i == 0 {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
