// Package i18n resolves the interface language and looks up translated messages.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

//go:embed locales/*.json
var files embed.FS

var fallback, active map[string]string

func load(lang string) map[string]string {
	b, err := files.ReadFile(path.Join("locales", lang+".json"))
	if err != nil {
		return nil
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		panic("i18n: broken locale " + lang + ": " + err.Error())
	}
	return m
}

// Init selects the language. An empty argument means detecting it from the environment.
func Init(lang string) {
	fallback = load("en")
	if lang == "" {
		lang = Detect()
	}
	active = load(normalize(lang))
	if active == nil {
		active = fallback
	}
}

// T returns the message for key, formatted with args.
func T(key string, args ...any) string {
	msg, ok := active[key]
	if !ok {
		if msg, ok = fallback[key]; !ok {
			return key
		}
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

// Detect reads NOSUB_LANG and the usual POSIX variables, then asks the OS.
func Detect() string {
	for _, v := range []string{"NOSUB_LANG", "LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		if s := os.Getenv(v); s != "" && s != "C" && s != "POSIX" {
			return s
		}
	}
	return systemLocale()
}

// normalize turns values like pt_BR.UTF-8 or zh-Hans-CN into a locale file name.
func normalize(s string) string {
	s, _, _ = strings.Cut(s, ":")
	s, _, _ = strings.Cut(s, ".")
	s, _, _ = strings.Cut(s, "@")
	s, _, _ = strings.Cut(strings.ReplaceAll(s, "_", "-"), "-")
	return strings.ToLower(s)
}

// Languages lists the available locale codes.
func Languages() []string {
	entries, _ := files.ReadDir("locales")
	var out []string
	for _, e := range entries {
		out = append(out, strings.TrimSuffix(e.Name(), ".json"))
	}
	sort.Strings(out)
	return out
}
